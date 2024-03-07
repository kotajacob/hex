package cache

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

type Comment struct {
	ID        int `json:"id"`
	Content   template.HTML
	Published time.Time  `json:"published"`
	Updated   *time.Time `json:"updated"`
	Path      string     `json:"path"`

	CreatorName string
	Upvotes     int
	Children    []*Comment
}

type Comments []Comment

// Comments returns the comments associated with a given post.
// The cached version is returned if it exists, otherwise, they are fetched.
// There is no cache expiration. Calling the Post method will check the
// expiration of both the post and its comments, updating the cache as needed.
// It should be called before this method.
func (c *Cache) Comments(cli *hb.Client, postID int) (Comments, error) {
	c.comments.mutex.RLock()
	comments, ok := c.comments.cache[postID]
	c.comments.mutex.RUnlock()
	if ok {
		return comments, nil
	}

	err := c.fetchComments(cli, postID)
	if err != nil {
		return comments, err
	}

	c.comments.mutex.RLock()
	comments = c.comments.cache[postID]
	c.comments.mutex.RUnlock()
	return comments, nil
}

// fetchComments retrieves all comments for a post making as many requests as
// needed.
func (c *Cache) fetchComments(cli *hb.Client, postID int) error {
	c.infoLog.Println("fetching comments for post:", postID)
	var all Comments
	page := 1
	limit := 50 // 50 seems to be the max we can request.
	for {
		views, resp, err := cli.CommentList(
			context.Background(),
			page,
			limit,
			postID,
			hb.CommentSortTypeOld,
		)
		if err != nil || views == nil {
			return fmt.Errorf(
				"failed fetching comments for post %v: %v resp: %v",
				postID,
				err,
				resp,
			)
		}
		if len(views.Comments) == 0 {
			break
		}

		for _, view := range views.Comments {
			var buf bytes.Buffer
			if err := markdown.Convert(
				[]byte(view.Comment.Content),
				&buf,
			); err != nil {
				return err
			}
			content := template.HTML(buf.Bytes())

			all = append(all, Comment{
				ID:        view.Comment.ID,
				Content:   content,
				Path:      view.Comment.Path,
				Published: view.Comment.Published,
				Updated:   view.Comment.Updated,

				CreatorName: view.Creator.Name,
				Upvotes:     view.Counts.Upvotes,
			})
		}
		if len(views.Comments) < limit {
			break
		}
		page += 1
	}

	c.comments.mutex.Lock()
	c.comments.cache[postID] = tree(all)
	c.comments.mutex.Unlock()
	return nil
}

func tree(all Comments) Comments {
	root := new(Comment)
	for _, comment := range all {
		root.addChild(comment)
	}

	var roots Comments
	for _, child := range root.Children {
		roots = append(roots, *child)
	}
	return roots
}

func (parent *Comment) addChild(child Comment) {
	var id int
	path := strings.Split(child.Path, ".")
	if len(path) > 1 {
		var err error
		id, err = strconv.Atoi(path[len(path)-2])
		if err != nil {
			panic("AHHH")
		}
	}

	if id == parent.ID {
		parent.Children = append(parent.Children, &child)
	}

	for _, c := range parent.Children {
		c.addChild(child)
	}
}
