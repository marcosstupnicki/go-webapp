package gowebapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLParam(t *testing.T) {
	webapp, err := NewWebApp("dummy-env", "8080")
	require.NoError(t, err)

	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("pong" + URLParam(r, "id")))
	}

	webapp.Get("/{id}", ping)

	// Test using the actual router
	req, _ := http.NewRequest("GET", "/value", nil)
	rr := httptest.NewRecorder()

	webapp.Router.ServeHTTP(rr, req)

	require.Equal(t, 200, rr.Code)
	require.Equal(t, "pongvalue", rr.Body.String())
}
