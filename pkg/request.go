package gowebapp

import (
	"github.com/go-chi/chi"
	"net/http"
)

// URLParam returns the url parameter from a http.Request object.
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}