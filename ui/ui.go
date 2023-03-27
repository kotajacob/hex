package ui

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"
)

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
		files := []string{"base.tmpl"}
		files = append(files, partials...)
		files = append(files, page)

		ts, err := template.ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}
