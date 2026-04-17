package gowebapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLParam(t *testing.T) {
	webapp := mustNew(t, "dummy-env", "8080")

	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("pong" + URLParam(r, "id")))
	}

	// Route registration moved from the WebApp helper to the Router in v2.
	webapp.Router.Get("/{id}", ping)

	req, _ := http.NewRequest("GET", "/value", nil)
	rr := httptest.NewRecorder()

	webapp.Router.ServeHTTP(rr, req)

	require.Equal(t, 200, rr.Code)
	require.Equal(t, "pongvalue", rr.Body.String())
}
