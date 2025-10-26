package gowebapp

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter_Method(t *testing.T) {
	webapp := NewWebApp("test", "8080")

	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("."))
	}

	// Test that we can register routes directly
	webapp.Get("/ping", ping)
	webapp.Post("/ping", ping)

	// Verify the router is working
	require.NotNil(t, webapp.Router)
	require.NotNil(t, webapp.Router.mux)
}
