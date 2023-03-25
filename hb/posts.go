package hb

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/yuin/goldmark"
)

// Post is a single post on hexbear.
type Post struct {
	ID          int           `json:"id"`
	Name        string        `json:"name"`
	URL         string        `json:"url"`
	Body        template.HTML `json:"body"`
	CommunityID int           `json:"community_id"`
	Comments    []Comment     `json:"comments"`

	// Image is a URL to a header image. During processing, if the URL contains
	// an image hosted on hexbear, we set this field and set the URL to blank.
	Image string
}

type PostList struct {
	Posts []Post `json:"posts"`
}

type PostComments struct {
	Post     Post     `json:"post"`
	Comments Comments `json:"comments"`
}

type Comment struct {
	ID      int    `json:"id"`
	Content string `json:"content"`
}
type Comments []Comment

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

	// Check if the URL is an image.
	if strings.HasPrefix(p.URL, "https://www.hexbear.net/pictrs/image/") {
		p.Image = p.URL
		p.URL = ""
	}
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
) (*PostComments, *http.Response, error) {
	path := fmt.Sprintf(
		"post?id=%d",
		id,
	)
	pc := new(PostComments)
	rsp, err := c.Do(ctx, path, pc)
	if err != nil {
		return pc, rsp, err
	}

	if err := processPost(&pc.Post); err != nil {
		return pc, rsp, err
	}

	return pc, rsp, err
}
