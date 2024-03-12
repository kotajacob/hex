package main

import (
	"net/http"
	"strconv"

	"git.sr.ht/~kota/hex/cache"
	"github.com/julienschmidt/httprouter"
)

type communityPage struct {
	CSPNonce string
	Name     string
	Page     int
	Posts    []cache.Post
}

// community handles displaying the lists of posts for a specific community.
func (app *application) community(w http.ResponseWriter, r *http.Request) {
	pageNum := 1
	q := r.URL.Query()
	if q.Has("page") {
		var err error
		pageNum, err = strconv.Atoi(q.Get("page"))
		if err != nil {
			app.notFound(w)
			return
		}
	}

	params := httprouter.ParamsFromContext(r.Context())
	name := params.ByName("name")
	community, err := app.cache.Community(app.client, name)
	if err != nil {
		app.clientError(w, http.StatusNotFound) // TODO: Handle server errors vs notFound error.
		return
	}

	page, err := app.cache.CommunityPosts(app.client, name, pageNum)
	if err != nil {
		app.serverError(w, err)
		return
	}
	var posts []cache.Post
	for _, id := range page.PostIDs {
		p, err := app.cache.Post(app.client, id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		posts = append(posts, p)
	}

	app.render(w, http.StatusOK, "community.tmpl", communityPage{
		CSPNonce: app.cspNonce,
		Name:     community.Name,
		Page:     pageNum,
		Posts:    posts,
	})
}

type communitiesPage struct {
	CSPNonce    string
	Communities []cache.Community
}

// communities handles displaying the community list page.
func (app *application) communities(w http.ResponseWriter, r *http.Request) {
	cms, err := app.cache.Communities()
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, http.StatusOK, "communities.tmpl", communitiesPage{
		CSPNonce:    app.cspNonce,
		Communities: cms,
	})
}
