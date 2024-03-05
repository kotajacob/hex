package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~kota/hex/cache"
)

const baseTMPL = "base.tmpl"

//go:embed "base.tmpl" "partials" "pages" "images"
var EFS embed.FS

func Templates() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	partials, err := fs.Glob(EFS, "partials/*.tmpl")
	if err != nil {
		panic(err)
	}

	pages, err := fs.Glob(EFS, "pages/*.tmpl")
	if err != nil {
		panic(err)
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := []string{baseTMPL}
		files = append(files, partials...)
		files = append(files, page)

		ts, err := template.New(baseTMPL).
			Funcs(template.FuncMap{"Timestamp": Timestamp}).
			ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
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
			b.WriteString(Since(*v.Updated, time.Now()))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(Since(v.Published, time.Now()))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	case *cache.Comment:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(Since(*v.Updated, time.Now()))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(Since(v.Published, time.Now()))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	case cache.Post:
		var b strings.Builder
		if v.Updated != nil {
			b.WriteString("<time title=\"")
			b.WriteString(v.Updated.String())
			b.WriteString("\">")
			b.WriteString(Since(*v.Updated, time.Now()))
		} else {
			b.WriteString("<time title=\"")
			b.WriteString(v.Published.String())
			b.WriteString("\">")
			b.WriteString(Since(v.Published, time.Now()))
		}
		b.WriteString("</time>")
		return template.HTML(b.String())
	default:
		return ""
	}
}

func Since(start, end time.Time) string {
	elapsed := end.Sub(time.Time(start))
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
