package gowebapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter_Method(t *testing.T) {
	webapp, err := NewWebApp("test", "8080")
	require.NoError(t, err)

	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("."))
	}

	// Test that we can register routes directly
	webapp.Get("/ping", ping)
	webapp.Post("/ping", ping)

	// Simple request to ensure ServeHTTP is wired
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	webapp.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)

	// Verify the router is working
	require.NotNil(t, webapp.Router)
	require.NotNil(t, webapp.Router.mux)
}
