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

type PostsResponse struct {
	Posts []Post `json:"posts"`
}

// GetPosts fetches Posts.
// Hexbear itself seems to fetch 40 posts when loading the home page.
// TODO: sort method and type.
func (c *Client) GetPosts(
	ctx context.Context,
	page int,
	limit int,
) (*PostsResponse, *http.Response, error) {
	path := fmt.Sprintf(
		"post/list?page=%d&limit=%d&sort=Active&type_=All",
		page,
		limit,
	)
	posts := new(PostsResponse)
	rsp, err := c.Do(ctx, path, posts)
	return posts, rsp, err
}
