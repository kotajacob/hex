package cache

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/url"
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

type Comments []*Comment

// Comments returns the comments associated with a given post.
// The cached version is returned if it exists, otherwise, they are fetched.
// There is no cache expiration. Calling the Post method will check the
// expiration of both the post and its comments, updating the cache as needed.
// It should be called before this method.
func (c *Cache) Comments(cli *hb.Client, postID int) (Comments, error) {
	comments, ok := c.comments.get(postID)
	if ok {
		return comments, nil
	}

	err := c.fetchComments(cli, postID)
	if err != nil {
		return comments, err
	}

	comments, _ = c.comments.get(postID)
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
			hb.CommentSortTypeHot,
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
			content, err := c.processMarkdown(view.Comment.Content)
			if err != nil {
				return err
			}

			all = append(all, &Comment{
				ID:        view.Comment.ID,
				Content:   content,
				Path:      view.Comment.Path,
				Published: view.Comment.Published,
				Updated:   view.Comment.Updated,

				CreatorName: processCreatorName(view.Creator),
				Upvotes:     view.Counts.Upvotes,
			})
		}
		if len(views.Comments) < limit {
			break
		}
		page += 1
	}

	c.comments.set(postID, (tree(all)))
	return nil
}

func tree(all Comments) Comments {
	root := new(Comment)
	root.addAll(all, 1)

	// Return a slice of comments instead of a root node.
	var roots Comments
	for _, child := range root.Children {
		roots = append(roots, child)
	}
	return roots
}

// addAll attempts to recursively all given child comments to the parent comment.
//
// The list is not sorted by parents, instead, any children which were not
// added in the first round are repeated until they can be added.
// This _could_ take a very long time in theory, but in practice it normally
// takes 1-5 iterations with each iteration being smaller than the previous.
//
// If more than 50 iterations would take place the process is cancelled and
// whatever remaining problem comments are dropped. These comments probably did
// not have a parent.
//
// This approach means we do not need to re-sort the comments after insertion.
func (parent *Comment) addAll(children []*Comment, iteration int) int {
	// Prevent an out of control stack overflow.
	if iteration > 50 {
		return iteration
	}

	var missing []*Comment
	for _, child := range children {
		found := parent.addChild(child)
		if !found {
			missing = append(missing, child)
		}
	}

	if len(missing) > 0 {
		parent.addAll(missing, iteration+1)
	}
	return iteration
}

// addChild attempts to add a comment to a tree rooted at parent.
// The tree will be traversed to find the appropriate parent. The success of
// this operation is returned: if the parent could be found.
func (parent *Comment) addChild(child *Comment) bool {
	var parentID int
	path := strings.Split(child.Path, ".")
	if len(path) > 1 {
		var err error
		parentID, err = strconv.Atoi(path[len(path)-2])
		if err != nil {
			// This should never happen! The comment's path is malformed so we
			// silently drop the comment returning success.
			return true
		}
	}

	// Is this comment a child of the current parent?
	if parentID == parent.ID {
		parent.Children = append(parent.Children, child)
		return true
	}

	// The comment is a child of a different parent.
	// Recursively call addChild on all siblings.
	for _, c := range parent.Children {
		if c.addChild(child) {
			return true
		}
	}

	// The comment's parent was not found in the tree.
	return false
}

func (c *Cache) processMarkdown(s string) (template.HTML, error) {
	var html template.HTML
	var buf bytes.Buffer
	if err := c.markdown.Convert(
		[]byte(s),
		&buf,
	); err != nil {
		return html, err
	}
	return template.HTML(c.emojiReplacer.Replace(buf.String())), nil
}

func processCreatorName(person hb.Person) string {
	if person.Local {
		if person.DisplayName != "" {
			return person.DisplayName
		}
		return person.Name
	}
	u, err := url.Parse(person.ActorID)
	if err != nil {
		return person.Name
	}
	return person.Name + "@" + u.Hostname()
}
