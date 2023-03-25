package main

import "net/http"

// serverError writes to the error log and writes a StatusInternalServerError to
// the client.
func (app *application) serverError(w http.ResponseWriter, err error) {
	app.errLog.Println(err)
	http.Error(
		w,
		http.StatusText(http.StatusInternalServerError),
		http.StatusInternalServerError,
	)
}

// clientError returns a particular error code and message to the client.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// notFound returns a 404 to the client.
func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}
