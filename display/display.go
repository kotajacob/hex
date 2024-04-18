package display

import (
	"html/template"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~kota/hex/cache"
	"git.sr.ht/~kota/hex/hb"
)

// NextPage renders a URL for the next page button in a community listing.
func NextPage(page int, sort string) string {
	var u url.URL
	q := u.Query()

	s := hb.ParseSortType(sort)
	if s != hb.DefaultSortType {
		q.Add("sort", strings.ToLower(string(s)))
	}

	q.Add("page", strconv.Itoa(page+1))
	return "?" + q.Encode()
}

// PrevPage renders a URL for the next page button in a community listing.
func PrevPage(page int, sort string) string {
	var u url.URL
	q := u.Query()

	s := hb.ParseSortType(sort)
	if s != hb.DefaultSortType {
		q.Add("sort", strings.ToLower(string(s)))
	}

	if page > 0 {
		q.Add("page", strconv.Itoa(page-1))
	}
	return "?" + q.Encode()
}

// Timestamp implements a fancy HTML timstamp renderer for the hb types.
func Timestamp(i interface{}) template.HTML {
	switch v := i.(type) {
	case cache.Comment:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(Since(*v.Updated))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(Since(v.Published))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	case *cache.Comment:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(Since(*v.Updated))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(Since(v.Published))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	case cache.Post:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(Since(*v.Updated))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(Since(v.Published))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	default:
		return ""
	}
}

func Since(start time.Time) string {
	elapsed := time.Now().Sub(time.Time(start))
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

func Date(t time.Time) string {
	return t.Format("January 2, 2006")
}
