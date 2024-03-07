package cache

import (
	"log"
	"sync"
	"time"

	"git.sr.ht/~kota/hex/hb"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

var markdown = goldmark.New(
	goldmark.WithExtensions(
		extension.NewLinkify(
			extension.WithLinkifyAllowedProtocols([][]byte{
				[]byte("http:"),
				[]byte("https:"),
			}),
		),
	),
)

type communityCache struct {
	mutex *sync.RWMutex
	cache map[int]Community
}

type postCache struct {
	mutex *sync.RWMutex
	cache map[int]Post
}

type commentCache struct {
	mutex *sync.RWMutex
	cache map[int]Comments
}

// The Cache is used to serve all requests. When available and fresh cached
// data is used, but fresh data will be fetched as needed.
type Cache struct {
	infoLog *log.Logger

	// home is a list of posts on the homepage.
	home Home

	// communities is a mapping of community IDs to the data representing them.
	communities communityCache

	// posts is a mapping of post IDs to the data representing them.
	posts postCache

	// comments is a mapping of post IDs to the root comments for that post.
	comments commentCache
}

// Initialize the cache and populate the communities and home page.
func Initialize(cli *hb.Client, infoLog *log.Logger) (*Cache, error) {
	c := new(Cache)
	c.communities.mutex = new(sync.RWMutex)
	c.communities.cache = make(map[int]Community)
	c.posts.mutex = new(sync.RWMutex)
	c.posts.cache = make(map[int]Post)
	c.comments.mutex = new(sync.RWMutex)
	c.comments.cache = make(map[int]Comments)
	c.infoLog = infoLog

	err := c.fetchCommunities(cli)
	if err != nil {
		return nil, err
	}

	err = c.fetchHome(cli)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// expired returns if a time is older than the duration.
func expired(t time.Time, d time.Duration) bool {
	return time.Since(t) > d
}
