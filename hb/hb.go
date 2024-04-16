// hb is a library for reading the lemmy v3 HTTP API.
package hb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// BaseURL is the default URL for the hexbear API.
const BaseURL = "https://www.hexbear.net/api/v3/"

// Client used for hb.
type Client struct {
	HTTPClient *http.Client
	BaseURL    *url.URL

	debugLog *log.Logger
}

// NewClient constructs a client using http.DefaultClient and the default
// base URL. The returned client is ready for use.
func NewClient(baseURL string, debugLog *log.Logger) (*Client, error) {
	var c Client
	u, err := url.Parse(baseURL)
	if err != nil {
		return &c, err
	}
	c.HTTPClient = http.DefaultClient
	c.BaseURL = u
	c.debugLog = debugLog
	return &c, nil
}

// StatusError is returned when a bad response code is received from the API.
type StatusError struct {
	Code int
}

var _ error = StatusError{}

func (e StatusError) Error() string {
	return fmt.Sprintf("bad response status code: %d", e.Code)
}

// Do sends an API request and returns the API response.
// The API response is JSON decoded and stored in the value pointed to by v, or
// returned as an error if an API error has occurred.
// If v is nil, and no error happens, the response is returned as is.
func (c *Client) Do(
	ctx context.Context,
	u *url.URL,
	v interface{},
) (*http.Response, error) {
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}
	req = req.WithContext(ctx)

	c.debugLog.Println("requesting:", u.String())
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, StatusError{Code: resp.StatusCode}
	}

	switch v := v.(type) {
	case nil:
	default:
		decErr := json.NewDecoder(resp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty response body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return resp, err
}

// ListingType is used when filtering a listing.
type ListingType string

const (
	ListingTypeAll        ListingType = "All"
	ListingTypeLocal      ListingType = "Local"
	ListingTypeSubscribed ListingType = "Subscribed"
)

// SortType is used when requesting sorted listings.
type SortType string

const (
	SortTypeActive         SortType = "Active"
	SortTypeHot            SortType = "Hot"
	SortTypeNew            SortType = "New"
	SortTypeOld            SortType = "Old"
	SortTypeTopDay         SortType = "TopDay"
	SortTypeTopWeek        SortType = "TopWeek"
	SortTypeTopMonth       SortType = "TopMonth"
	SortTypeTopYear        SortType = "TopYear"
	SortTypeTopAll         SortType = "TopAll"
	SortTypeMostComments   SortType = "MostComments"
	SortTypeNewComments    SortType = "NewComments"
	SortTypeTopHour        SortType = "TopHour"
	SortTypeTopSixHour     SortType = "TopSixHour"
	SortTypeTopTwelveHour  SortType = "TopTwelveHour"
	SortTypeTopThreeMonths SortType = "TopThreeMonths"
	SortTypeTopSixMonths   SortType = "TopSixMonths"
	SortTypeTopNineMonths  SortType = "TopNineMonths"
)

func ParseSortType(s string) SortType {
	s = strings.ToLower(s)
	switch s {
	case "active":
		return SortTypeActive
	case "hot":
		return SortTypeHot
	case "new":
		return SortTypeNew
	case "old":
		return SortTypeOld
	case "topday":
		return SortTypeTopDay
	case "topweek":
		return SortTypeTopWeek
	case "topmonth":
		return SortTypeTopMonth
	case "topyear":
		return SortTypeTopYear
	case "topall":
		return SortTypeTopAll
	case "mostcomments":
		return SortTypeMostComments
	case "newcomments":
		return SortTypeNewComments
	case "tophour":
		return SortTypeTopHour
	case "topsixhour":
		return SortTypeTopSixHour
	case "toptwelvehour":
		return SortTypeTopTwelveHour
	case "topthreemonths":
		return SortTypeTopThreeMonths
	case "topsixmonths":
		return SortTypeTopSixMonths
	case "topninemonths":
		return SortTypeTopNineMonths
	default:
		return SortTypeActive
	}
}

// CommentSortType is used when requesting sorted comments.
type CommentSortType string

const (
	CommentSortTypeHot CommentSortType = "Hot"
	CommentSortTypeTop CommentSortType = "Top"
	CommentSortTypeNew CommentSortType = "New"
	CommentSortTypeOld CommentSortType = "Old"
)

func ParseCommentSortType(s string) CommentSortType {
	s = strings.ToLower(s)
	switch s {
	case "hot":
		return CommentSortTypeHot
	case "top":
		return CommentSortTypeTop
	case "new":
		return CommentSortTypeNew
	case "old":
		return CommentSortTypeOld
	default:
		return CommentSortTypeHot
	}
}
