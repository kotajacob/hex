package main

import (
	"context"
	"fmt"

	"git.sr.ht/~kota/hex/hb"
)

type cache struct {
	communities map[int]hb.Community
	posts       map[int]hb.Post
}

func populateCache(cli *hb.Client) (*cache, error) {
	communities := make(map[int]hb.Community)
	posts := make(map[int]hb.Post)

	cms, resp, err := cli.CommunityList(context.TODO())
	if err != nil || cms == nil {
		return nil, fmt.Errorf(
			"failing getting communities: %v resp: %v",
			err,
			resp,
		)
	}
	for _, cm := range cms.Communities {
		communities[cm.ID] = cm
	}

	ps, resp, err := cli.PostList(context.TODO(), 1, 40)
	if err != nil || ps == nil {
		return nil, fmt.Errorf(
			"failing getting font-page posts: %v resp: %v",
			err,
			resp,
		)
	}
	var allPosts []int
	for _, p := range ps.Posts {
		posts[p.ID] = p
		allPosts = append(allPosts, p.ID)
	}
	// Create a fake "all" community to store an ordered list of posts.
	communities[0] = hb.Community{
		ID:    0,
		Name:  "all",
		Title: "All the posts from hexbear",
		Posts: allPosts,
	}

	return &cache{
		communities: communities,
		posts:       posts,
	}, nil
}
