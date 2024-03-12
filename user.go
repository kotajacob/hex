package main

import (
	"net/http"
	"time"

	"git.sr.ht/~kota/hex/cache"
	"github.com/julienschmidt/httprouter"
)

type userPage struct {
	CSPNonce     string
	Name         string
	Bio          string
	CommentCount int
	PostCount    int
	Created      time.Time

	Posts []cache.Post
}

// user handles displaying information for a specific user.
func (app *application) user(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	name := params.ByName("name")
	user, err := app.cache.Person(app.client, name)
	if err != nil {
		app.clientError(w, http.StatusNotFound) // TODO: Handle server errors vs notFound error.
		return
	}

	var posts []cache.Post
	for _, id := range user.PostIDs {
		p, err := app.cache.Post(app.client, id)
		if err != nil {
			app.serverError(w, err)
			return
		}
		posts = append(posts, p)
	}

	app.render(w, http.StatusOK, "user.tmpl", userPage{
		CSPNonce:     app.cspNonce,
		Name:         user.DisplayName,
		Bio:          user.Bio,
		CommentCount: user.CommentCount,
		PostCount:    user.PostCount,
		Created:      user.Published,

		Posts: posts,
	})
}
