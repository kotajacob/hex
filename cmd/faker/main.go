package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var root string

func postList(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(filepath.Join(root, "/post/list.json"))
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	w.Write(data)
}

func communityList(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(filepath.Join(root, "/community/list.json"))
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	w.Write(data)
}

func post(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	if id == "" {
	}
	data, err := os.ReadFile(filepath.Join(root, "/post/"+id+".json"))
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
		return
	}
	w.Write(data)
}

func main() {
	flag.StringVar(&root, "root", "cmd/faker/testdata", "testdata directory")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/post", post)
	mux.HandleFunc("/post/list", postList)
	mux.HandleFunc("/community/list", communityList)

	log.Println("starting faker on :4001")
	err := http.ListenAndServe(":4001", mux)
	log.Fatal(err)
}
