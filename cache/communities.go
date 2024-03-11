package cache

import (
	"context"
	"fmt"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

type Community struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`

	PostIDs []int
	Fetched time.Time
}

// HomePage represents all the posts on a particular page number of the Home
// view.
type HomePage struct {
	PostIDs []int
	Fetched time.Time
}

// Community returns a community.
// The cached version is returned if it exists, otherwise, all communities are
// fetched and updated.
func (c *Cache) Community(cli *hb.Client, id int) (Community, error) {
	comm, ok := c.communities.get(id)
	if ok {
		return comm, nil
	}

	err := c.fetchCommunities(cli)
	if err != nil {
		return comm, err
	}

	comm, _ = c.communities.get(id)
	return comm, nil
}

// Communities returns a list of all cached communities.
func (c *Cache) Communities() ([]Community, error) {
	return c.communities.getAll(), nil
}

// fetchCommunities retrieves all local hexbear communities.
func (c *Cache) fetchCommunities(cli *hb.Client) error {
	c.infoLog.Println("fetching communities")
	now := time.Now()

	page := 1
	limit := 50 // 50 seems to be the max we can request.
	for {
		views, resp, err := cli.CommunityList(
			context.Background(),
			page,
			limit,
			hb.ListingTypeLocal,
		)
		if err != nil || views == nil {
			return fmt.Errorf(
				"failing getting communities: %v resp: %v",
				err,
				resp,
			)
		}
		if len(views.Communities) == 0 {
			break
		}
		for _, view := range views.Communities {
			c.communities.set(view.Community.ID, Community{
				ID:          view.Community.ID,
				Name:        view.Community.Name,
				Title:       view.Community.Title,
				Description: view.Community.Description,

				Fetched: now,
			})
		}
		if len(views.Communities) < limit {
			break
		}
		page += 1
	}

	return nil
}
