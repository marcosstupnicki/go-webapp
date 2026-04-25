package gowebapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter_Method(t *testing.T) {
	webapp := mustNew(t, "test", "8080")

	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("."))
	}

	webapp.Get("/ping", ping)
	webapp.Post("/ping", ping)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()
	webapp.ServeHTTP(rr, req)
	require.Equal(t, 200, rr.Code)
	require.NotNil(t, webapp.Router)
}
