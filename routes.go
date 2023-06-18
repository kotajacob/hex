package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

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
	MOTD     template.HTML
	Posts    []hb.Post
}

// home handles displaying the home page.
func (app *application) home(w http.ResponseWriter, r *http.Request) {
	community, ok := app.cache.communities[0]
	if !ok || expired(app.cache.communitiesFetched[0]) {
		var err error
		community, err = app.fetchHome()
		if err != nil {
			app.serverError(w, err)
			return
		}
	}
	var posts []hb.Post
	for _, id := range community.Posts {
		posts = append(posts, app.cache.posts[id])
	}

	app.render(w, http.StatusOK, "home.tmpl", homePage{
		CSPNonce: app.cspNonce,
		MOTD:     hb.GetMOTD(),
		Posts:    posts,
	})
}

// fetchHome does a live fetch of the home page data.
func (app *application) fetchHome() (hb.Community, error) {
	ps, resp, err := app.cli.PostList(context.TODO(), 1, 40)
	if err != nil || ps == nil {
		return hb.Community{}, fmt.Errorf(
			"failing getting font-page posts: %v resp: %v",
			err,
			resp,
		)
	}
	var allPosts []int
	now := time.Now()
	for _, p := range ps.Posts {
		app.cache.posts[p.ID] = p
		app.cache.postsFetched[p.ID] = now
		allPosts = append(allPosts, p.ID)
	}

	// Create a fake "all" community for home page.
	allCommunity := hb.Community{
		ID:    0,
		Name:  "all",
		Title: "All the posts from hexbear",
		Posts: allPosts,
	}
	app.cache.communities[0] = allCommunity
	app.cache.communitiesFetched[0] = now
	return allCommunity, nil
}

type postPage struct {
	CSPNonce  string
	Post      hb.Post
	Comments  []hb.Comment
	Community hb.Community
}

// post handles requests for displaying a post's comment page.
func (app *application) post(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil || id < 1 {
		app.notFound(w)
		return
	}

	var post hb.Post
	comments, ok := app.cache.comments[id]
	if !ok || expired(app.cache.commentsFetched[id]) {
		var err error
		post, comments, err = app.fetchPost(id)
		if errors.As(err, &hb.StatusError{}) {
			app.notFound(w)
			return
		} else if err != nil {
			app.serverError(w, err)
			return
		}
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

// fetchPost does a live fetch of a post and its comments.
func (app *application) fetchPost(id int) (hb.Post, hb.Comments, error) {
	pc, _, err := app.cli.Post(context.TODO(), id)
	if err != nil {
		return pc.Post, pc.Comments, err
	}
	now := time.Now()
	post := pc.Post
	comments := pc.Comments
	app.cache.posts[id] = post
	app.cache.postsFetched[id] = now
	app.cache.comments[id] = comments
	app.cache.commentsFetched[id] = now
	return pc.Post, pc.Comments, nil
}

type communitiesPage struct {
	CSPNonce    string
	Communities map[int]hb.Community
}

// communities handles displaying the community list page.
func (app *application) communities(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "communities.tmpl", communitiesPage{
		CSPNonce:    app.cspNonce,
		Communities: app.cache.communities,
	})
}

// ppb does exactly what you'd expect.
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
