package hb

import (
	"context"
	"net/http"
	"strconv"
)

// Community is a single community on hexbear.
type Community struct {
	ID          int    `json:"id"`
	Local       bool   `json:"local"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CommunityView represents a Community and additional metadata.
type CommunityView struct {
	Community Community `json:"community"`
}

// CommunityListResp is a list of CommunityViews.
type CommunityListResp struct {
	Communities []CommunityView `json:"communities"`
}

// CommunityList fetches a slice of all communities.
func (c *Client) CommunityList(
	ctx context.Context,
	page int,
	limit int,
	listingType ListingType,
) (*CommunityListResp, *http.Response, error) {
	u := c.BaseURL.JoinPath("community/list")
	q := u.Query()
	if page != 0 {
		q.Add("page", strconv.Itoa(page))
	}
	if limit != 0 {
		q.Add("limit", strconv.Itoa(limit))
	}
	if listingType != "" {
		q.Add("type_", string(listingType))
	}
	u.RawQuery = q.Encode()

	communities := new(CommunityListResp)
	resp, err := c.Do(ctx, u, communities)
	return communities, resp, err
}
