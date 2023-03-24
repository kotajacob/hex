package hb

import (
	"context"
	"fmt"
	"net/http"
)

// Post is a single post on hexbear.
type Post struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
	Body string `json:"body"`
}
type PostList struct {
	Posts []Post `json:"posts"`
}

// PostLst fetches a slice of posts.
// Hexbear itself seems to fetch 40 posts when loading the home page.
func (c *Client) PostList(
	// TODO: support sort method and type.
	ctx context.Context,
	page int,
	limit int,
) (*PostList, *http.Response, error) {
	path := fmt.Sprintf(
		"post/list?page=%d&limit=%d&sort=Active&type_=All",
		page,
		limit,
	)
	posts := new(PostList)
	rsp, err := c.Do(ctx, path, posts)
	return posts, rsp, err
}
