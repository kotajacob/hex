package hb

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/yuin/goldmark"
)

// Post is a single post on hexbear.
type Post struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	Body        template.HTML `json:"body"`
	CommunityID int           `json:"community_id"`
}
type PostList struct {
	Posts []Post `json:"posts"`
}

// processPost makes all nessesary modifications to the Post after it's fetched.
func processPost(p *Post) error {
	// Process the post's body with goldmark.
	var buf bytes.Buffer
	if err := goldmark.Convert(
		[]byte(p.Body),
		&buf,
	); err != nil {
		return err
	}
	p.Body = template.HTML(buf.Bytes())
	return nil
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
	postList := new(PostList)
	rsp, err := c.Do(ctx, path, postList)
	if err != nil {
		return postList, rsp, err
	}

	for i := range postList.Posts {
		p := &postList.Posts[i]
		if err := processPost(p); err != nil {
			return postList, rsp, err
		}
	}
	return postList, rsp, err
}

// Post fetches a single post by ID.
func (c *Client) Post(
	ctx context.Context,
	id int,
) (*Post, *http.Response, error) {
	path := fmt.Sprintf(
		"post?id=%d",
		id,
	)
	post := new(Post)
	rsp, err := c.Do(ctx, path, post)
	if err != nil {
		return post, rsp, err
	}

	if err := processPost(post); err != nil {
		return post, rsp, err
	}

	return post, rsp, err
}
