package main

import (
	"fmt"
	"io"
	"net/http"

	"git.sr.ht/~kota/hex/hb"
	"git.sr.ht/~kota/hex/ui"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/communities", app.communities)
	mux.HandleFunc("/ppb", app.ppb)
	return app.recoverPanic(app.logRequest(app.secureHeaders(mux)))
}

type listPage struct {
	CSPNonce string
	MOTD     string
	Posts    []hb.Post
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	tsName := "home.tmpl"
	ts, ok := app.templates[tsName]
	if !ok {
		app.serverError(w, fmt.Errorf(
			"the template %s is missing",
			tsName,
		))
		return
	}

	var posts []hb.Post
	for _, id := range app.cache.communities[0].Posts {
		posts = append(posts, app.cache.posts[id])
	}

	err := ts.ExecuteTemplate(w, "base", listPage{
		CSPNonce: app.cspNonce,
		MOTD:     hb.GetMOTD(),
		Posts:    posts,
	})
	if err != nil {
		app.serverError(w, err)
	}
}

type communitiesPage struct {
	CSPNonce    string
	Communities map[int]hb.Community
}

func (app *application) communities(w http.ResponseWriter, r *http.Request) {
	tsName := "communities.tmpl"
	ts, ok := app.templates[tsName]
	if !ok {
		app.serverError(w, err)
		return
	}

	err := ts.ExecuteTemplate(w, "base", communitiesPage{
		CSPNonce:    app.cspNonce,
		Communities: app.cache.communities,
	})
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) ppb(w http.ResponseWriter, r *http.Request) {
	f, err := ui.EFS.Open("images/ppb.jpg")
	if err != nil {
		app.serverError(w, fmt.Errorf("failed to open ppb.jpg: %v", err))
		return
	}
	data, err := io.ReadAll(f)
	if err != nil {
		app.serverError(w, fmt.Errorf("failed to read ppb.jpg: %v", err))
		return
	}
	w.Write(data)
}
