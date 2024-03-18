package cache

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"time"

	"git.sr.ht/~kota/hex/hb"
)

const PERSON_TTL = time.Minute * 40

type Person struct {
	ActorID     string        `json:"actor_id"` // URL for home server.
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Bio         template.HTML `json:"bio"`
	Local       bool          `json:"local"`
	Published   time.Time     `json:"published"`
	Updated     time.Time     `json:"updated"`

	CommentCount int
	PostCount    int
	PostIDs      []int
	Fetched      time.Time
}

// Person returns a given Person.
// The cached version is returned if it exists and has not expired, otherwise,
// they are fetched. The user's posts are also retrieved as part of this
// request.
func (c *Cache) Person(cli *hb.Client, name string) (Person, error) {
	person, ok := c.persons.get(name)
	if !ok || expired(person.Fetched, PERSON_TTL) {
		err := c.fetchPerson(cli, name)
		if err != nil {
			return person, err
		}
		person, _ = c.persons.get(name)
	}
	return person, nil
}

// fetchPerson retrieves a person along with their posts.
func (c *Cache) fetchPerson(cli *hb.Client, name string) error {
	c.infoLog.Println("fetching person:", name)

	pr, resp, err := cli.Person(context.Background(), 0, name)
	if err != nil || pr == nil {
		return fmt.Errorf("failed fetching person: %v resp: %v", err, resp)
	}

	var postIDs []int
	for _, postView := range pr.Posts {
		err = c.storePost(postView)
		if err != nil {
			c.errLog.Println("failed to add post", postView.Post.ID, err)
		}
		postIDs = append(postIDs, postView.Post.ID)
	}

	bio, err := c.processMarkdown(pr.PersonView.Person.Bio)
	if err != nil {
		return err
	}

	c.persons.set(name, Person{
		ActorID:     pr.PersonView.Person.ActorID,
		Name:        pr.PersonView.Person.Name,
		DisplayName: processCreatorName(pr.PersonView.Person),
		Bio:         bio,
		Local:       pr.PersonView.Person.Local,
		Published:   pr.PersonView.Person.Published,
		Updated:     pr.PersonView.Person.Updated,

		CommentCount: pr.PersonView.Counts.CommentCount,
		PostCount:    pr.PersonView.Counts.PostCount,
		PostIDs:      postIDs,
		Fetched:      time.Now(),
	})
	return nil
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

func processCreatorURL(person hb.Person) string {
	u, err := url.Parse(person.ActorID)
	if err != nil || person.Local {
		return "/u/" + person.Name
	}
	return u.String()
}
