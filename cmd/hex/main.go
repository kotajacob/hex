package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"

	"git.sr.ht/~kota/hex/hb"
	"git.sr.ht/~kota/hex/ui"
)

type application struct {
	infoLog *log.Logger
	errLog  *log.Logger

	cache     *cache
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
	errLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	templates, err := ui.Templates()
	if err != nil {
		errLog.Fatal(err)
	}

	cli := hb.NewClient(*hbURL)
	cache, err := populateCache(cli)
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
