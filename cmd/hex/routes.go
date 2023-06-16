package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"git.sr.ht/~kota/hex/hb"
	"git.sr.ht/~kota/hex/ui"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/post/:id", app.post)
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

type postPage struct {
	CSPNonce  string
	Post      hb.Post
	Comments  []hb.Comment
	Community hb.Community
}

func (app *application) post(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	var post hb.Post
	comments, ok := app.cache.comments[id]
	if !ok {
		// Do a live fetch of the post and comments.
		pc, _, err := app.cli.Post(context.TODO(), id)
		if errors.As(err, &hb.StatusError{}) {
			app.notFound(w)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}
		post = pc.Post
		comments = pc.Comments
		app.cache.posts[id] = post
		app.cache.comments[id] = comments
	} else {
		post = app.cache.posts[id]
	}

	community := app.cache.communities[post.CommunityID]
	app.render(w, http.StatusOK, "post.tmpl", postPage{
		CSPNonce:  app.cspNonce,
		Post:      post,
		Comments:  comments,
		Community: community,
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
