package gowebapp

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

// mustNew is a test helper that calls New and fails the test on error.
func mustNew(t *testing.T, env, port string, opts ...Option) *WebApp {
	t.Helper()
	app, err := New(env, port, opts...)
	require.NoError(t, err)
	return app
}

func TestNew(t *testing.T) {
	webapp, err := New("dummy-env", "8080")
	require.NoError(t, err)
	require.Equal(t, Scope{Environment: "dummy-env"}, webapp.Scope)
	require.Equal(t, "8080", webapp.Port)
	require.NotNil(t, webapp.Router)
}

func TestWebApp_Group(t *testing.T) {
	webapp := mustNew(t, "test", "8080")

	webapp.Group(func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("test"))
		})
	})

	require.NotNil(t, webapp.Router)
	require.NotNil(t, webapp.Router.mux)
}
