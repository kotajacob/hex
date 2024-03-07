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

type Home struct {
	PostIDs []int
	Fetched time.Time
}

// Community returns a community.
// The cached version is returned if it exists, otherwise, all communities are
// fetched and updated.
func (c *Cache) Community(cli *hb.Client, id int) (Community, error) {
	c.communities.mutex.RLock()
	comm, ok := c.communities.cache[id]
	c.communities.mutex.RUnlock()
	if ok {
		return comm, nil
	}

	err := c.fetchCommunities(cli)
	if err != nil {
		return comm, err
	}

	c.communities.mutex.RLock()
	comm = c.communities.cache[id]
	c.communities.mutex.RUnlock()
	return comm, nil
}

// Communities returns a list of all cached communities.
func (c *Cache) Communities() ([]Community, error) {
	var cms []Community

	c.communities.mutex.RLock()
	for _, cm := range c.communities.cache {
		cms = append(cms, cm)
	}
	c.communities.mutex.RUnlock()

	return cms, nil
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
			c.communities.mutex.Lock()
			c.communities.cache[view.Community.ID] = Community{
				ID:          view.Community.ID,
				Name:        view.Community.Name,
				Title:       view.Community.Title,
				Description: view.Community.Description,

				Fetched: now,
			}
			c.communities.mutex.Unlock()
		}
		if len(views.Communities) < limit {
			break
		}
		page += 1
	}

	return nil
}
