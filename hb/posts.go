package hb

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// Post is a single post.
type Post struct {
	ID          int        `json:"id"`
	Local       bool       `json:"local"`
	Name        string     `json:"name"`
	URL         string     `json:"url"`
	Body        string     `json:"body"`
	CommunityID int        `json:"community_id"`
	Published   time.Time  `json:"published"`
	Updated     *time.Time `json:"updated"`
	CreatorID   int        `json:"creator_id"`
}

// PostAggregates is aggregated scores for a post.
type PostAggregates struct {
	Score             int  `json:"score"`
	Comments          int  `json:"comments"`
	Upvotes           int  `json:"upvotes"`
	Downvotes         int  `json:"downvotes"`
	FeaturedCommunity bool `json:"featured_community"`
	FeaturedLocal     bool `json:"featured_local"`
}

// PostView represents a Post and additional metadata.
type PostView struct {
	Post      Post           `json:"post"`
	Creator   Person         `json:"creator"`
	Community Community      `json:"community"`
	Counts    PostAggregates `json:"counts"`
}

// PostListResp is the response from PostList.
type PostListResp struct {
	Posts []PostView `json:"posts"`
}

// PostResp is the response from Post.
type PostResp struct {
	PostView PostView `json:"post_view"`
}

// PostList fetches a slice of posts.
func (c *Client) PostList(
	ctx context.Context,
	communityID int,
	page int,
	limit int,
	sortType SortType,
	listingType ListingType,
) (*PostListResp, *http.Response, error) {
	u := c.BaseURL.JoinPath("post/list")
	q := u.Query()
	if communityID != 0 {
		q.Add("community_id", strconv.Itoa(communityID))
	}
	if page != 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if limit != 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	if sortType != "" {
		q.Add("sort", string(sortType))
	}
	if listingType != "" {
		q.Add("type_", string(listingType))
	}
	u.RawQuery = q.Encode()

	posts := new(PostListResp)
	resp, err := c.Do(ctx, u, posts)
	return posts, resp, err
}

// Post fetches a single post.
func (c *Client) Post(
	ctx context.Context,
	id int,
) (*PostResp, *http.Response, error) {
	u := c.BaseURL.JoinPath("post")
	q := u.Query()
	if id != 0 {
		q.Add("id", strconv.Itoa(id))
	}
	u.RawQuery = q.Encode()

	post := new(PostResp)
	resp, err := c.Do(ctx, u, post)
	return post, resp, err
}
