package gowebapp

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestURLParam(t *testing.T) {
	webapp := NewWebApp("dummy-env")

	ping := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("pong" + URLParam(r, "iidd")))
	}

	webapp.Get("/{id}", ping)

	var tests = []struct {
		name string
		request *http.Request
		key string
		expectedParam string

	}{
		{
			name: "URLParam Ok",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "/value", nil)

				return req
			}(),
			key: "id",
			expectedParam: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlParam := URLParam(tt.request, tt.key)
			require.Equal(t, tt.expectedParam, urlParam)
		})
	}
}