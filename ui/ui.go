package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
)

//go:embed "base.tmpl" "pages" "images"
var EFS embed.FS

func Templates() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(EFS, "pages/*.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		files := []string{
			"base.tmpl",
			page,
		}

		ts, err := template.ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
