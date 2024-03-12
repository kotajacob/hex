package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net/http"

	"git.sr.ht/~kota/hex/files"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	emojiFS, err := fs.Sub(files.EFS, "emoji")
	if err != nil {
		panic("unable to locate emoji folder")
	}
	fileServer := http.FileServer(http.FS(emojiFS))
	router.Handler(http.MethodGet, "/pictrs/image/*filepath", http.StripPrefix("/pictrs/image", fileServer))

	router.HandlerFunc(http.MethodGet, "/", app.home)
	router.HandlerFunc(http.MethodGet, "/post/:id", app.post)
	router.HandlerFunc(http.MethodGet, "/c/:name", app.community)
	router.HandlerFunc(http.MethodGet, "/u/:name", app.user)
	router.HandlerFunc(http.MethodGet, "/communities", app.communities)
	router.HandlerFunc(http.MethodGet, "/ppb", app.ppb)
	return app.recoverPanic(app.logRequest(app.secureHeaders(router)))
}

func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data interface{},
) {
	ts, ok := app.templates[page]
	if !ok {
		app.serverError(w, fmt.Errorf(
			"the template %s is missing",
			page,
		))
		return
	}

	buf := new(bytes.Buffer)

	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.serverError(w, err)
	}

	w.WriteHeader(status)
	buf.WriteTo(w)
}

// ppb does exactly what you'd expect.
func (app *application) ppb(w http.ResponseWriter, r *http.Request) {
	f, err := files.EFS.Open("images/ppb.jpg")
	if err != nil {
		app.serverError(w, fmt.Errorf("failed to open ppb.jpg: %v", err))
		return
	}
	data, err := io.ReadAll(f)
	if err != nil {
		app.serverError(w, fmt.Errorf("failed to read ppb.jpg: %v", err))
		return
	}
	w.Write(data)
}
