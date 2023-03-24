package main

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
)

// cspNonce securely generates a 128bit base64 encoded number.
func cspNonce() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	return base64.RawStdEncoding.EncodeToString(b), err
}

// secureHeaders is a middleware which adds strict CSP and other headers.
func (app *application) secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := cspNonce()
		if err != nil {
			log.Println(err)
			http.Error(
				w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
			return
		}
		w.Header().Set(
			"Content-Security-Policy",
			"default-src 'none'; style-src 'nonce-"+nonce+"'",
		)
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")
		app.cspNonce = nonce

		next.ServeHTTP(w, r)
	})
}

// logRequest is a middleware that prints each request to the info log.
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf(
			"%s - %s %s %s",
			r.RemoteAddr,
			r.Proto,
			r.Method,
			r.URL.RequestURI(),
		)
		next.ServeHTTP(w, r)
	})
}
