package files

import (
	"embed"
	"html/template"
	"io/fs"
	"path/filepath"

	"git.sr.ht/~kota/hex/display"
)

const baseTMPL = "base.tmpl"

//go:embed "base.tmpl" "partials" "pages" "images" "emoji"
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
			Funcs(template.FuncMap{
				"Increment": display.Increment,
				"Decrement": display.Decrement,
				"Timestamp": display.Timestamp,
			}).
			ParseFS(EFS, files...)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}
	return cache, nil
}

func Emojis() []string {
	emojis, err := EFS.ReadDir("emoji")
	if err != nil {
		panic("failed to read embeded emoji folder")
	}

	var oldnew []string
	for _, emoji := range emojis {
		name := emoji.Name()
		oldnew = append(oldnew, "https://hexbear.net/pictrs/image/"+name)
		oldnew = append(oldnew, "/pictrs/image/"+name)
		oldnew = append(oldnew, "https://www.hexbear.net/pictrs/image/"+name)
		oldnew = append(oldnew, "/pictrs/image/"+name)
	}
	return oldnew
}
