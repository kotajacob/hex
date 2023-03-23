package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	"git.sr.ht/~kota/hex/ui"
)

func home(w http.ResponseWriter, r *http.Request) {
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

	err = ts.Execute(w, "hello world")
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)

	log.Println("starting server on", *addr)
	err := http.ListenAndServe(*addr, mux)
	log.Fatal(err)
}
