package hb

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

type Comment struct {
	ID        int        `json:"id"`
	Content   string     `json:"content"`
	CreatorID int        `json:"creator_id"`
	Deleted   bool       `json:"deleted"`
	Local     bool       `json:"local"`
	Path      string     `json:"path"`
	PostID    int        `json:"post_id"`
	Published time.Time  `json:"published"`
	Updated   *time.Time `json:"updated"`
}

type CommentAggregates struct {
	Score      int `json:"score"`
	ChildCount int `json:"child_count"`
	Upvotes    int `json:"upvotes"`
	Downvotes  int `json:"downvotes"`
}

type CommentView struct {
	Comment            Comment           `json:"comment"`
	Community          Community         `json:"community"`
	Counts             CommentAggregates `json:"counts"`
	Post               Post              `json:"post"`
	Creator            Person            `json:"creator"`
	CreatorIsAdmin     bool              `json:"creator_is_admin"`
	CreatorIsModerator bool              `json:"creator_is_moderator"`
}

type CommentListResp struct {
	Comments []CommentView `json:"comments"`
}

// CommentList is used to get some or all comments associated with a post.
func (c *Client) CommentList(
	ctx context.Context,
	page int,
	limit int,
	postID int,
	sortType CommentSortType,
) (*CommentListResp, *http.Response, error) {
	u := c.BaseURL.JoinPath("comment/list")
	q := u.Query()
	if page != 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if limit != 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	if postID != 0 {
		q.Add("post_id", strconv.Itoa(postID))
	}
	if sortType != "" {
		q.Add("sort", string(sortType))
	}
	u.RawQuery = q.Encode()

	comments := new(CommentListResp)
	resp, err := c.Do(ctx, u, comments)
	return comments, resp, err
}
