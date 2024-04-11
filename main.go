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

// DOMAIN is the domain name this server is assumed to be hosted under.
// It's only really used for replacing hexbear links within the site.
// It can be overwritten with a launch flag.
const DOMAIN = "https://diethex.net"

type application struct {
	infoLog *log.Logger
	errLog  *log.Logger

	client    *hb.Client
	cache     *cache.Cache
	templates map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	hbURL := flag.String("hb", hb.BaseURL, "hexbear baseURL")
	domain := flag.String("domain", DOMAIN, "domain name for link replacement")
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
	linkReplacer := newLinkReplacer(*domain)

	cli, err := hb.NewClient(*hbURL, debugLog)
	if err != nil {
		errLog.Fatalf(
			"failed creating hexbear client %v",
			err,
		)
	}
	cache, err := cache.Initialize(
		cli,
		infoLog,
		errLog,
		markdown,
		emojiReplacer,
		linkReplacer,
	)
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
