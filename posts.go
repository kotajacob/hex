package main

import (
	"net/http"
	"strconv"

	"git.sr.ht/~kota/hex/cache"
	"git.sr.ht/~kota/hex/hb"
	"github.com/julienschmidt/httprouter"
)

type postPage struct {
	CSPNonce    string
	Post        cache.Post
	Comments    []*cache.Comment
	CommentSort string
}

// post handles requests for displaying a post's comment page.
func (app *application) post(w http.ResponseWriter, r *http.Request) {
	sort := hb.ParseCommentSortType(r.URL.Query().Get("sort"))

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

	comments, err := app.cache.Comments(app.client, id, sort)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, http.StatusOK, "post.tmpl", postPage{
		CSPNonce:    nonce(r.Context()),
		Post:        post,
		Comments:    comments.Comments,
		CommentSort: string(sort),
	})
}
