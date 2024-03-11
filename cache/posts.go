package cache

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

type Post struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Body        template.HTML
	CommunityID int        `json:"community_id"`
	Published   time.Time  `json:"published"`
	Updated     *time.Time `json:"updated"`

	CreatorName   string
	CommunityName string
	Image         string
	Upvotes       int
	CommentCount  int
	Fetched       time.Time
}

// Post returns a given post.
// The cached version is returned if it exists and has not expired, otherwise,
// they are fetched. If the post is fetched its comments are also fetched.
func (c *Cache) Post(cli *hb.Client, id int) (Post, error) {
	post, ok := c.posts.get(id)
	if !ok || expired(post.Fetched, time.Minute*20) {
		err := c.fetchPost(cli, id)
		if err != nil {
			return post, err
		}
		post, _ = c.posts.get(id)
	}
	return post, nil
}

// fetchPost retrieves a given post and all of its comments.
func (c *Cache) fetchPost(cli *hb.Client, postID int) error {
	c.infoLog.Println("fetching post:", postID)

	pr, resp, err := cli.Post(context.Background(), postID)
	if err != nil || pr == nil {
		return fmt.Errorf(
			"failing fetching post: %v resp: %v",
			err,
			resp,
		)
	}

	err = c.storePost(pr.PostView)
	if err != nil {
		return err
	}
	return c.fetchComments(cli, postID)
}

// Home returns the home page.
// The cached version is returned if it exists and has not expired, otherwise,
// they are fetched fresh. If the posts are fetched their comments are NOT
// fetched.
func (c *Cache) Home(cli *hb.Client, page int) (HomePage, error) {
	home, ok := c.home.get(page)
	if ok && !expired(home.Fetched, time.Minute*20) {
		return home, nil
	}
	err := c.fetchHome(cli, page)
	home, _ = c.home.get(page)
	return home, err
}

// fetchHome retrieves all of the posts needed for the home page.
func (c *Cache) fetchHome(cli *hb.Client, page int) error {
	c.infoLog.Println("fetching home posts page:", page)
	now := time.Now()

	limit := 50 // 50 seems to be the max we can request.
	home := HomePage{
		Fetched: now,
	}
	views, resp, err := cli.PostList(
		context.Background(),
		0,
		page,
		limit,
		hb.SortTypeActive,
		hb.ListingTypeLocal,
	)
	if err != nil || views == nil {
		return fmt.Errorf(
			"failing fetching home posts page: %v: err %v: resp: %v",
			page,
			err,
			resp,
		)
	}

	for _, view := range views.Posts {
		c.storePost(view)
		home.PostIDs = append(home.PostIDs, view.Post.ID)
	}

	c.home.set(page, home)
	return nil
}

// storePost converts an hb.PostView into a Post and stores it in the cache.
func (c *Cache) storePost(view hb.PostView) error {
	url := view.Post.URL
	var image string
	if strings.HasPrefix(view.Post.URL, "https://hexbear.net/pictrs/image/") {
		image = url
		url = ""
	}

	body, err := c.processMarkdown(view.Post.Body)
	if err != nil {
		return err
	}

	c.posts.set(view.Post.ID, Post{
		ID:          view.Post.ID,
		Name:        view.Post.Name,
		URL:         url,
		Body:        body,
		CommunityID: view.Post.CommunityID,
		Published:   view.Post.Published,
		Updated:     view.Post.Updated,

		CreatorName:   processCreatorName(view.Creator),
		CommunityName: view.Community.Name,
		Image:         image,
		Upvotes:       view.Counts.Upvotes,
		CommentCount:  view.Counts.Comments,
		Fetched:       time.Now(),
	})
	return nil
}
