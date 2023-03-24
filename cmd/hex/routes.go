package main

import (
	"html/template"
	"log"
	"net/http"

	"git.sr.ht/~kota/hex/hb"
	"git.sr.ht/~kota/hex/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	return mux
}

type listPage struct {
	MOTD  string
	Posts []hb.Post
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

	var posts []hb.Post
	for _, id := range app.cache.communities[0].Posts {
		posts = append(posts, app.cache.posts[id])
	}

	err = ts.Execute(w, listPage{
		MOTD:  hb.GetMOTD(),
		Posts: posts,
	})
	if err != nil {
		log.Println(err)
		http.Error(
			w,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}
