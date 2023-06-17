// hb is a library for reading the hexbear v1 HTTP API.
package hb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BaseURL is the default URL for the hexbear API.
const BaseURL = "https://www.hexbear.net/api/v1/"

// Client used for hb.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewClient constructs a client using http.DefaultClient and the default
// base URL. The returned client is ready for use.
func NewClient(baseURL string) *Client {
	return &Client{
		HTTPClient: http.DefaultClient,
		BaseURL:    baseURL,
	}
}

// StatusError is returned when a bad responce code is received from the API.
type StatusError struct {
	Code int
}

var _ error = StatusError{}

func (e StatusError) Error() string {
	return fmt.Sprintf("bad responce status code: %d", e.Code)
}

// HBTime represents the time format used by the hexbear API.
// The format itself is semi-nonstandard and also will sometimes be present,
// but with the literal string "null".
type HBTime time.Time

func (h *HBTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return err
	}
	*h = HBTime(t)
	return nil
}

func (h HBTime) String() string {
	return time.Time(h).Format(time.DateTime)
}

func (h HBTime) Since(t time.Time) string {
	elapsed := t.Sub(time.Time(h))
	years := int(math.Floor(elapsed.Hours() / 24 / 365))
	days := int(math.Floor(elapsed.Hours() / 24))
	hours := int(math.Floor(elapsed.Hours()))
	minutes := int(math.Floor(elapsed.Minutes()))
	seconds := int(math.Floor(elapsed.Seconds()))
	switch {
	case years > 0:
		return strconv.Itoa(years) + " years ago"
	case days > 0:
		return strconv.Itoa(days) + " days ago"
	case hours > 0:
		return strconv.Itoa(hours) + " hours ago"
	case minutes > 0:
		return strconv.Itoa(minutes) + " minutes ago"
	case seconds > 0:
		return strconv.Itoa(seconds) + " seconds ago"
	default:
		return "0 seconds ago"
	}
}

// Do sends an API request and returns the API responce.
// The API responce is JSON decoded and stored in the value pointed to by v, or
// returned as an error if an API error has occured.
// If v is nil, and no error happens, the responce is returned as is.
func (c *Client) Do(
	ctx context.Context,
	path string,
	v interface{},
) (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %v", err)
	}
	req = req.WithContext(ctx)

	rsp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do request: %v", err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return nil, StatusError{Code: rsp.StatusCode}
	}

	switch v := v.(type) {
	case nil:
	default:
		decErr := json.NewDecoder(rsp.Body).Decode(v)
		if decErr == io.EOF {
			decErr = nil // ignore EOF errors caused by empty responce body
		}
		if decErr != nil {
			err = decErr
		}
	}
	return rsp, err
}

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }
