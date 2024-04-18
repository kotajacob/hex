package main

import (
	"net/http"
	"strconv"

	"git.sr.ht/~kota/hex/cache"
	"git.sr.ht/~kota/hex/hb"
)

// home handles displaying the home page.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
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
	sort := hb.ParseSortType(q.Get("sort"))

	page, err := app.cache.Home(app.client, pageNum, sort)
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
		CSPNonce: nonce(r.Context()),
		Message:  hb.GetMOTD(),
		Page:     pageNum,
		Posts:    posts,
		Sort:     string(sort),
	})
}
