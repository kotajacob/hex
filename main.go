package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"git.sr.ht/~kota/hex/cache"
	"git.sr.ht/~kota/hex/files"
	"git.sr.ht/~kota/hex/hb"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

type application struct {
	infoLog *log.Logger
	errLog  *log.Logger

	client    *hb.Client
	cache     *cache.Cache
	templates map[string]*template.Template

	// In order to use inline css, we need to set a randomly generated nonce
	// value for each request. This is set in our secureHeaders middleware and
	// then used in our base template.
	cspNonce string
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	hbURL := flag.String("hb", hb.BaseURL, "hexbear baseURL")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	debugLog := log.New(os.Stdout, "DEBUG ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	templates, err := files.Templates()
	if err != nil {
		errLog.Fatal(err)
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(
			extension.NewLinkify(
				extension.WithLinkifyAllowedProtocols([][]byte{
					[]byte("http:"),
					[]byte("https:"),
				}),
			),
			extension.Strikethrough,
		),
	)

	emojiReplacer := strings.NewReplacer(files.Emojis()...)

	cli, err := hb.NewClient(*hbURL, debugLog)
	if err != nil {
		errLog.Fatalf(
			"failed creating hexbear client %v",
			err,
		)
	}
	cache, err := cache.Initialize(cli, infoLog, markdown, emojiReplacer)
	if err != nil {
		errLog.Fatalf(
			"failed populating initial cache %v",
			err,
		)
	}
	app := &application{
		infoLog:   infoLog,
		errLog:    errLog,
		cache:     cache,
		client:    cli,
		templates: templates,
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errLog,
		Handler:  app.routes(),
	}

	infoLog.Println("starting server on", *addr)
	err = srv.ListenAndServe()
	errLog.Fatal(err)
}
