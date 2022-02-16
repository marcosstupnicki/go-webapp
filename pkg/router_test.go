package gowebapp

import (
	"net/http"
	"testing"
)

func TestRouter_Method(t *testing.T) {
	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("."))
	}
}
