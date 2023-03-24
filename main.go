package main

import (
	"context"
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

	posts *hb.PostsResponse
}

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	return mux
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	ts, err := template.ParseFS(ui.EFS, "main.tmpl")
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}

	err = ts.Execute(w, app.posts)
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}

func main() {
	addr := flag.String("addr", ":4000", "HTTP network address")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lshortfile)

	cli := hb.NewClient()
	posts, resp, err := cli.GetPosts(context.Background(), 1, 40)
	if err != nil {
		errLog.Fatalf(
			"failed fetching initial posts %v: response: %+v",
			err,
			resp,
		)
	}

	app := &application{
		infoLog: infoLog,
		errLog:  errLog,
		posts:   posts,
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
