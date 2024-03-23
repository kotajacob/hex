package cache

import (
	"context"
	"fmt"
	"html/template"
	"net/url"
	"strings"
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
		ActorID: pr.PersonView.Person.ActorID,
		Name:    pr.PersonView.Person.Name,
		DisplayName: processPersonName(
			pr.PersonView.Person,
			pr.PersonView.IsAdmin,
			false, // No concept of moderator on a person view page.
			false, // No concept of OP on a person view page.
		),
		Bio:       bio,
		Local:     pr.PersonView.Person.Local,
		Published: pr.PersonView.Person.Published,
		Updated:   pr.PersonView.Person.Updated,

		CommentCount: pr.PersonView.Counts.CommentCount,
		PostCount:    pr.PersonView.Counts.PostCount,
		PostIDs:      postIDs,
		Fetched:      time.Now(),
	})
	return nil
}

// processPersonName creates a DisplayName to store in cache for a person.
// Pronouns are displayed, then an [A] for admins, an [M] for mods, and [OP] to
// signify if the given user created the post being viewed. Remote users will
// instead have name@homeserver.
//
// In some contexts, such as a post listing, there is no concept of [OP], or on
// a person's page there's no concept of moderators. In these cases false
// should be passed for those booleans.
func processPersonName(person hb.Person, admin, mod, op bool) string {
	var s strings.Builder
	if person.Local {
		if person.DisplayName != "" {
			s.WriteString(person.DisplayName)
		} else {
			s.WriteString(person.Name)
		}
	} else {
		u, err := url.Parse(person.ActorID)
		if err != nil {
			s.WriteString(person.Name)
		} else {
			s.WriteString(person.Name + "@" + u.Hostname())
		}
	}

	if admin {
		s.WriteString(" [A]")
	}

	if mod {
		s.WriteString(" [M]")
	}

	if op {
		s.WriteString(" [OP]")
	}
	return s.String()
}

func processPersonURL(person hb.Person) string {
	u, err := url.Parse(person.ActorID)
	if err != nil || person.Local {
		return "/u/" + person.Name
	}
	return u.String()
}
