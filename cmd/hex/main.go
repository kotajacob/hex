package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"git.sr.ht/~kota/hex/hb"
)

type application struct {
	infoLog *log.Logger
	errLog  *log.Logger

	cache *cache
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	hbURL := flag.String("hb", hb.BaseURL, "hexbear baseURL")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	cli := hb.NewClient(*hbURL)
	cache, err := populateCache(cli)
	if err != nil {
		errLog.Fatalf(
			"failed populating initial cache %v",
			err,
		)
	}
	app := &application{
		infoLog: infoLog,
		errLog:  errLog,
		cache:   cache,
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
