package main

import (
	"context"
	"fmt"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

type cache struct {
	// communities is a mapping of community IDs to the data representing them.
	// Each community contains a list of post IDs which can be used to build
	// the front page for that community by looking up the posts.
	communities        map[int]hb.Community
	communitiesFetched map[int]time.Time

	// posts is a mapping of post IDs to the data representing them.
	// The posts themselves do not contain comments. Instead you can look them
	// up in the comments cache using the posts ID.
	posts        map[int]hb.Post
	postsFetched map[int]time.Time

	// comments is a mapping of post IDs to the root comments for that post.
	comments        map[int]hb.Comments
	commentsFetched map[int]time.Time
}

func initialCache(cli *hb.Client) (*cache, error) {
	communities := make(map[int]hb.Community)
	communitiesFetched := make(map[int]time.Time)
	posts := make(map[int]hb.Post)
	postsFetched := make(map[int]time.Time)
	comments := make(map[int]hb.Comments)
	commentsFetched := make(map[int]time.Time)
	now := time.Now()

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
		communitiesFetched[cm.ID] = now
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
		postsFetched[p.ID] = now
		allPosts = append(allPosts, p.ID)
	}

	// Create a fake "all" community for home page.
	communities[0] = hb.Community{
		ID:    0,
		Name:  "all",
		Title: "All the posts from hexbear",
		Posts: allPosts,
	}
	communitiesFetched[0] = now

	return &cache{
		communities:        communities,
		communitiesFetched: communitiesFetched,
		posts:              posts,
		postsFetched:       postsFetched,
		comments:           comments,
		commentsFetched:    commentsFetched,
	}, nil
}

// expired returns true if the time is older than 5 minutes.
func expired(t time.Time) bool {
	now := time.Now()
	elapsed := now.Sub(t)
	return elapsed.Minutes() > 5
}
