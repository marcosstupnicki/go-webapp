package gowebapp

import "net/http"

// DefaultSecurityHeaders returns the recommended baseline browser security
// headers for web applications.
func DefaultSecurityHeaders() map[string]string {
	return map[string]string{
		"Content-Security-Policy":   "default-src 'self';",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "geolocation=(), camera=()",
	}
}

// SecurityHeaders adds standard security response headers to every request.
func SecurityHeaders(headers map[string]string) func(http.Handler) http.Handler {
	copied := cloneHeaderMap(headers)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for key, value := range copied {
				w.Header().Set(key, value)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func securityHeadersMiddleware(headers map[string]string) func(http.Handler) http.Handler {
	return SecurityHeaders(headers)
}

func hasSecurityHeaders(headers map[string]string) bool {
	return len(headers) > 0
}
