package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"git.sr.ht/~kota/hex/hb"
	"git.sr.ht/~kota/hex/ui"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/communities", app.communities)
	router.HandlerFunc(http.MethodGet, "/ppb", app.ppb)
	return app.recoverPanic(app.logRequest(app.secureHeaders(router)))
}

func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data interface{},
) {
	ts, ok := app.templates[page]
	if !ok {
		app.serverError(w, fmt.Errorf(
			"the template %s is missing",
			page,
		))
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

type homePage struct {
	CSPNonce string
	MOTD     string
	Posts    []hb.Post
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	var posts []hb.Post
	for _, id := range app.cache.communities[0].Posts {
		posts = append(posts, app.cache.posts[id])
	}

	app.render(w, http.StatusOK, "home.tmpl", homePage{
		CSPNonce: app.cspNonce,
		MOTD:     hb.GetMOTD(),
		Posts:    posts,
	})
}

type communitiesPage struct {
	CSPNonce    string
	Communities map[int]hb.Community
}

func (app *application) communities(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "communities.tmpl", communitiesPage{
		CSPNonce:    app.cspNonce,
		Communities: app.cache.communities,
	})
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
