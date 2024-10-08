package cache

import (
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

const (
	POSTS_PER_PAGE = 50 // 50 seems to be the max we can request.
	PAGE_TTL       = time.Minute * 15
	POST_TTL       = time.Minute * 15
)

type Post struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	URL               string `json:"url"`
	Body              template.HTML
	CommunityID       int        `json:"community_id"`
	Published         time.Time  `json:"published"`
	Updated           *time.Time `json:"updated"`
	FeaturedLocal     bool       `json:"featured_local"`
	FeaturedCommunity bool       `json:"featured_community"`

	CreatorDisplayName string
	CreatorID          int
	CreatorURL         string
	CommunityName      string
	Image              string
	Upvotes            int
	CommentCount       int
	Fetched            time.Time
}

// Post returns a given post.
// The cached version is returned if it exists and has not expired, otherwise,
// they are fetched.
func (c *Cache) Post(cli *hb.Client, id int) (Post, error) {
	post, ok := c.posts.get(id)
	if !ok || expired(post.Fetched, POST_TTL) {
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
		return fmt.Errorf("failing fetching post: %v resp: %v", err, resp)
	}

	return c.storePost(pr.PostView)
}

// Home returns the home page.
// The cached version is returned if it exists and has not expired, otherwise,
// they are fetched fresh. If the posts are fetched their comments are NOT
// fetched.
func (c *Cache) Home(cli *hb.Client, page int, sort hb.SortType) (Page, error) {
	home, ok := c.home.get(page, sort)
	if ok && !expired(home.Fetched, PAGE_TTL) {
		return home, nil
	}
	err := c.fetchHome(cli, page, sort)
	home, _ = c.home.get(page, sort)
	return home, err
}

// fetchHome retrieves all of the posts needed for the home page.
func (c *Cache) fetchHome(cli *hb.Client, page int, sort hb.SortType) error {
	c.infoLog.Println("fetching home posts page:", page)
	now := time.Now()

	limit := POSTS_PER_PAGE
	home := Page{
		Fetched: now,
	}
	views, resp, err := cli.PostList(
		context.Background(),
		0,
		page,
		limit,
		sort,
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
		err = c.storePost(view)
		if err != nil {
			c.errLog.Println("failed to add post", view.Post.ID, err)
		}
		home.PostIDs = append(home.PostIDs, view.Post.ID)
	}

	c.home.set(page, sort, home)
	return nil
}

// CommunityPosts returns a page of posts within a community.
// The cached version is returned if it exists and has not expired, otherwise,
// they are fetched fresh. If the posts are fetched their comments are NOT
// fetched.
func (c *Cache) CommunityPosts(
	cli *hb.Client,
	communityName string,
	pageNum int,
	sort hb.SortType,
) (Page, error) {
	community, ok := c.communities.get(communityName)
	if !ok {
		err := c.fetchCommunities(cli)
		if err != nil {
		}
		community, _ = c.communities.get(communityName)
	}

	page, ok := community.get(pageNum, sort)
	if ok && !expired(page.Fetched, PAGE_TTL) {
		return page, nil
	}
	err := c.fetchCommunityPosts(cli, communityName, pageNum, sort)
	page, _ = community.get(pageNum, sort)
	return page, err
}

// fetchCommunityPosts retrieves all of the posts for a given page of a community.
func (c *Cache) fetchCommunityPosts(
	cli *hb.Client,
	communityName string,
	pageNum int,
	sort hb.SortType,
) error {
	community, ok := c.communities.get(communityName)
	if !ok {
		return fmt.Errorf("requested community %v does not exist", communityName)
	}
	c.infoLog.Printf("fetching %v posts page: %v\n", community.Name, pageNum)
	now := time.Now()

	limit := POSTS_PER_PAGE
	page := Page{
		Fetched: now,
	}
	views, resp, err := cli.PostList(
		context.Background(),
		community.ID,
		pageNum,
		limit,
		sort,
		hb.ListingTypeLocal,
	)
	if err != nil || views == nil {
		return fmt.Errorf(
			"failing fetching %v posts page: %v: err %v: resp: %v",
			community.Name,
			pageNum,
			err,
			resp,
		)
	}

	for _, view := range views.Posts {
		err = c.storePost(view)
		if err != nil {
			c.errLog.Println("failed to add post", view.Post.ID, err)
		}
		page.PostIDs = append(page.PostIDs, view.Post.ID)
	}

	community.set(pageNum, sort, page)
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
		ID:                view.Post.ID,
		Name:              view.Post.Name,
		URL:               url,
		Body:              body,
		CommunityID:       view.Post.CommunityID,
		Published:         view.Post.Published,
		Updated:           view.Post.Updated,
		FeaturedLocal:     view.Post.FeaturedLocal,
		FeaturedCommunity: view.Post.FeaturedCommunity,

		CreatorDisplayName: processPersonName(
			view.Creator,
			view.CreatorIsAdmin,
			view.CreatorIsModerator,
			false, // No need to mark them as OP when it's obvious.
		),
		CreatorID:     view.Creator.ID,
		CreatorURL:    processPersonURL(view.Creator),
		CommunityName: view.Community.Name,
		Image:         image,
		Upvotes:       view.Counts.Upvotes,
		CommentCount:  view.Counts.Comments,
		Fetched:       time.Now(),
	})
	return nil
}
