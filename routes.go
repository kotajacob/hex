package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strconv"
	"time"

	"git.sr.ht/~kota/hex/cache"
	"git.sr.ht/~kota/hex/files"
	"git.sr.ht/~kota/hex/hb"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	emojiFS, err := fs.Sub(files.EFS, "emoji")
	if err != nil {
		panic("unable to locate emoji folder")
	}
	fileServer := http.FileServer(http.FS(emojiFS))
	router.Handler(http.MethodGet, "/pictrs/image/*filepath", http.StripPrefix("/pictrs/image", fileServer))

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/post/:id", app.post)
	router.HandlerFunc(http.MethodGet, "/c/:name", app.community)
	router.HandlerFunc(http.MethodGet, "/u/:name", app.user)
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
	MOTD     template.HTML
	Page     int
	Posts    []cache.Post
}

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

	page, err := app.cache.Home(app.client, pageNum)
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

	app.render(w, http.StatusOK, "home.tmpl", homePage{
		CSPNonce: app.cspNonce,
		MOTD:     hb.GetMOTD(),
		Page:     pageNum,
		Posts:    posts,
	})
}

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

// ppb does exactly what you'd expect.
func (app *application) ppb(w http.ResponseWriter, r *http.Request) {
	f, err := files.EFS.Open("images/ppb.jpg")
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
