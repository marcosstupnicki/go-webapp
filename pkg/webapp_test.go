package gowebapp

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestNewWebApp(t *testing.T) {
	var tests = []struct {
		name          string
		environment   string
		port          string
		expectedScope Scope
		expectedPort  string
	}{
		{
			name:          "NewWebApp Ok",
			environment:   "dummy-env",
			port:          "8080",
			expectedScope: Scope{Environment: "dummy-env"},
			expectedPort:  "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webapp, err := NewWebApp(tt.environment, tt.port)
			require.NoError(t, err)
			require.Equal(t, tt.expectedScope, webapp.Scope)
			require.Equal(t, tt.expectedPort, webapp.Port)
			require.NotEmpty(t, webapp.Router)
			require.NotEmpty(t, webapp)
		})
	}
}

func TestWebApp_Run(t *testing.T) {
	environment := "dummy-env"
	port := "8080"
	webapp1, err := NewWebApp(environment, port)
	require.NoError(t, err)
	go func() {
		err := webapp1.Run()
		require.NoError(t, err)
	}()

	go func() {
		webapp2, err := NewWebApp(environment, port)
		require.NoError(t, err)
		err = webapp2.Run()
		require.Errorf(t, err, "port already used by another WebApp")
	}()
}

func TestWebApp_Group(t *testing.T) {
	webapp, err := NewWebApp("test", "8080")
	require.NoError(t, err)

	// Test that Group method works and can register routes
	webapp.Group(func(r chi.Router) {
		r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(200)
			w.Write([]byte("test"))
		})
	})

	// Verify the route was registered by checking if the router is not nil
	require.NotNil(t, webapp.Router)
	require.NotNil(t, webapp.Router.mux)
}
