package main

import (
	"fmt"
	"log"
	"net/http"

	"git.sr.ht/~kota/hex/hb"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/communities", app.communities)
	return mux
}

type listPage struct {
	MOTD  string
	Posts []hb.Post
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	tsName := "home.tmpl"
	ts, ok := app.templates[tsName]
	if !ok {
		app.errLog.Println(fmt.Errorf(
			"the template %s is missing",
			tsName,
		))
		http.NotFound(w, r)
		return
	}

	var posts []hb.Post
	for _, id := range app.cache.communities[0].Posts {
		posts = append(posts, app.cache.posts[id])
	}

	err := ts.ExecuteTemplate(w, "base", listPage{
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

type communitiesPage struct {
	Communities map[int]hb.Community
}

func (app *application) communities(w http.ResponseWriter, r *http.Request) {
	tsName := "communities.tmpl"
	ts, ok := app.templates[tsName]
	if !ok {
		app.errLog.Println(fmt.Errorf(
			"the template %s is missing",
			tsName,
		))
		http.NotFound(w, r)
		return
	}

	err := ts.ExecuteTemplate(w, "base", communitiesPage{
		Communities: app.cache.communities,
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
