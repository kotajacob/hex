package main

import (
	"net/http"
	"strconv"

	"git.sr.ht/~kota/hex/cache"
	"github.com/julienschmidt/httprouter"
)

type postPage struct {
	CSPNonce string
	Post     cache.Post
	Comments []*cache.Comment
}

// post handles requests for displaying a post's comment page.
func (app *application) post(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	post, err := app.cache.Post(app.client, id)
	if err != nil { // TODO: Handle notFound error.
		app.serverError(w, err)
		return
	}

	comments, err := app.cache.Comments(app.client, id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, http.StatusOK, "post.tmpl", postPage{
		CSPNonce: app.cspNonce,
		Post:     post,
		Comments: comments,
	})
}
