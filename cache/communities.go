package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

type Community struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`

	// pages contains all the posts on a particular page for a Community.
	mutex *sync.RWMutex
	pages map[int]Page
}

// get attempts to get a page from the mapping of pages in a community.
func (c Community) get(num int) (Page, bool) {
	c.mutex.RLock()
	page, ok := c.pages[num]
	c.mutex.RUnlock()
	return page, ok
}

// set attempts to store a page in a community's page mapping.
func (c Community) set(num int, page Page) {
	c.mutex.Lock()
	c.pages[num] = page
	c.mutex.Unlock()
}

// A Page contains all the posts on a particular page.
type Page struct {
	PostIDs []int
	Fetched time.Time
}

// Community returns a Community by name.
// The cached version is returned if it exists, otherwise, all communities are
// fetched and updated.
// This does not fetch posts within this community.
func (c *Cache) Community(cli *hb.Client, name string) (Community, error) {
	comm, ok := c.communities.get(name)
	if ok {
		return comm, nil
	}

	err := c.fetchCommunities(cli)
	comm, ok = c.communities.get(name)
	if !ok && err == nil {
		err = fmt.Errorf("community %v does not exist", name)
	}
	return comm, err
}

// Communities returns a list of all cached communities.
// This does not fetch posts within these communities.
func (c *Cache) Communities() ([]Community, error) {
	return c.communities.getAll(), nil
}

// fetchCommunities retrieves all local hexbear communities.
// This does not fetch posts within these communities.
func (c *Cache) fetchCommunities(cli *hb.Client) error {
	c.infoLog.Println("fetching communities")

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
			c.communities.set(view.Community.Name, Community{
				ID:          view.Community.ID,
				Name:        view.Community.Name,
				Title:       view.Community.Title,
				Description: view.Community.Description,

				pages: make(map[int]Page),
				mutex: new(sync.RWMutex),
			})
		}
		if len(views.Communities) < limit {
			break
		}
		page += 1
	}

	return nil
}
