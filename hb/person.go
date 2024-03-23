package hb

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// Person is a single user.
type Person struct {
	ActorID     string    `json:"actor_id"` // URL for home server.
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	Local       bool      `json:"local"`
	Published   time.Time `json:"published"`
	Updated     time.Time `json:"updated"`
}

// PersonAggregates is aggregated scores for a person.
type PersonAggregates struct {
	CommentCount int `json:"comment_count"`
	PostCount    int `json:"post_count"`
}

// PersonView represents a Person along with some aggregate information about
// them.
type PersonView struct {
	Counts  PersonAggregates `json:"counts"`
	Person  Person           `json:"person"`
	IsAdmin bool             `json:"is_admin"`
}

// PersonResp is the response from Person.
type PersonResp struct {
	PersonView PersonView    `json:"person_view"`
	Comments   []CommentView `json:"comments"`
	Posts      []PostView    `json:"posts"`
}

// Person fetches information about a single user.
func (c *Client) Person(
	ctx context.Context,
	id int,
	name string,
) (*PersonResp, *http.Response, error) {
	u := c.BaseURL.JoinPath("user")
	q := u.Query()
	if id != 0 {
		q.Add("person_id", strconv.Itoa(id))
	}
	if name != "" {
		q.Add("username", name)
	}
	u.RawQuery = q.Encode()

	person := new(PersonResp)
	resp, err := c.Do(ctx, u, person)
	return person, resp, err
}
