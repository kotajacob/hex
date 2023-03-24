package hb

import (
	"context"
	"fmt"
	"net/http"
)

// Community is a single community on hexbear.
type Community struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// Posts is an orderd slice of post IDs in this community using default
	// sorting. This is not a real field returned from hexbear.
	Posts []int
}
type CommunityList struct {
	Communities []Community `json:"communities"`
}

// CommunityList fetches a slice of all communities.
func (c *Client) CommunityList(
	ctx context.Context,
) (*CommunityList, *http.Response, error) {
	path := fmt.Sprintf("community/list?sort=TopAll&limit=1000")
	communities := new(CommunityList)
	rsp, err := c.Do(ctx, path, communities)
	return communities, rsp, err
}
