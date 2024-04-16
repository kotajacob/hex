package cache

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"git.sr.ht/~kota/hex/hb"
	"github.com/yuin/goldmark"
)

// The Cache is used to serve all requests. When available and fresh cached
// data is used, but fresh data will be fetched as needed.
type Cache struct {
	infoLog *log.Logger
	errLog  *log.Logger

	markdown      goldmark.Markdown
	emojiReplacer *strings.Replacer
	linkReplacer  *strings.Replacer

	// home is a mapping of page_number:sorting_method to lists of posts.
	home homeCache

	// communities is a mapping of community names to information about that
	// community.
	// The communities themselves contain a mapping of pages to lists of posts.
	communities communityCache

	// posts is a mapping of post IDs to the data representing them.
	posts postCache

	// comments is a mapping of post_IDs:sorting_method to the root comments
	// for that post.
	comments commentCache

	// persons is a mapping of usernames to information about that person.
	persons personCache
}

type homeCache struct {
	mutex *sync.RWMutex
	cache map[string]Page
}

func newHomeCache() homeCache {
	var c homeCache
	c.mutex = new(sync.RWMutex)
	c.cache = make(map[string]Page)
	return c
}

func (c homeCache) get(num int, sort hb.SortType) (Page, bool) {
	c.mutex.RLock()
	key := strconv.Itoa(num) + ":" + string(sort)
	home, ok := c.cache[key]
	c.mutex.RUnlock()
	return home, ok
}

func (c homeCache) set(num int, sort hb.SortType, home Page) {
	c.mutex.Lock()
	key := strconv.Itoa(num) + ":" + string(sort)
	c.cache[key] = home
	c.mutex.Unlock()
}

type communityCache struct {
	mutex *sync.RWMutex
	cache map[string]Community
}

func newCommunityCache() communityCache {
	var c communityCache
	c.mutex = new(sync.RWMutex)
	c.cache = make(map[string]Community)
	return c
}

func (c communityCache) get(name string) (Community, bool) {
	c.mutex.RLock()
	community, ok := c.cache[name]
	c.mutex.RUnlock()
	return community, ok
}

func (c communityCache) getAll() []Community {
	var cms []Community
	c.mutex.RLock()
	for _, cm := range c.cache {
		cms = append(cms, cm)
	}
	c.mutex.RUnlock()
	return cms
}

func (c communityCache) set(name string, community Community) {
	c.mutex.Lock()
	c.cache[name] = community
	c.mutex.Unlock()
}

type postCache struct {
	mutex *sync.RWMutex
	cache map[int]Post
}

func newPostCache() postCache {
	var c postCache
	c.mutex = new(sync.RWMutex)
	c.cache = make(map[int]Post)
	return c
}

func (c postCache) get(id int) (Post, bool) {
	c.mutex.RLock()
	post, ok := c.cache[id]
	c.mutex.RUnlock()
	return post, ok
}

func (c postCache) set(id int, post Post) {
	c.mutex.Lock()
	c.cache[id] = post
	c.mutex.Unlock()
}

type commentCache struct {
	mutex *sync.RWMutex
	cache map[string]PostComments
}

func newCommentCache() commentCache {
	var c commentCache
	c.mutex = new(sync.RWMutex)
	c.cache = make(map[string]PostComments)
	return c
}

func (c commentCache) get(id int, sort hb.CommentSortType) (PostComments, bool) {
	c.mutex.RLock()
	key := strconv.Itoa(id) + ":" + string(sort)
	comments, ok := c.cache[key]
	c.mutex.RUnlock()
	return comments, ok
}

func (c commentCache) set(id int, sort hb.CommentSortType, comments PostComments) {
	c.mutex.Lock()
	key := strconv.Itoa(id) + ":" + string(sort)
	c.cache[key] = comments
	c.mutex.Unlock()
}

type personCache struct {
	mutex *sync.RWMutex
	cache map[string]Person
}

func newPersonCache() personCache {
	var c personCache
	c.mutex = new(sync.RWMutex)
	c.cache = make(map[string]Person)
	return c
}

func (c personCache) get(name string) (Person, bool) {
	c.mutex.RLock()
	persons, ok := c.cache[name]
	c.mutex.RUnlock()
	return persons, ok
}

func (c personCache) set(name string, persons Person) {
	c.mutex.Lock()
	c.cache[name] = persons
	c.mutex.Unlock()
}

// Initialize the cache and populate the communities and home page.
func Initialize(
	cli *hb.Client,
	infoLog *log.Logger,
	errLog *log.Logger,
	markdown goldmark.Markdown,
	emojiReplacer *strings.Replacer,
	linkReplacer *strings.Replacer,
) (*Cache, error) {
	c := new(Cache)
	c.infoLog = infoLog
	c.errLog = errLog

	c.home = newHomeCache()
	c.communities = newCommunityCache()
	c.posts = newPostCache()
	c.comments = newCommentCache()
	c.persons = newPersonCache()

	c.markdown = markdown
	c.emojiReplacer = emojiReplacer
	c.linkReplacer = linkReplacer

	err := c.fetchCommunities(cli)
	if err != nil {
		return nil, err
	}

	err = c.fetchHome(cli, 1, hb.SortTypeActive)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// expired returns if a time is older than the duration.
func expired(t time.Time, d time.Duration) bool {
	return time.Since(t) > d
}
