package gowebapp

import (
	"net/http"
	"strings"
)

// RequireJSONContentType rejects POST/PATCH/PUT requests whose Content-Type
// is present but not application/json. Requests with no Content-Type header
// are allowed through so handlers can decide whether a body is required.
func RequireJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
				_ = RespondWithError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds standard security response headers to every request.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}
