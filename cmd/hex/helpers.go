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
