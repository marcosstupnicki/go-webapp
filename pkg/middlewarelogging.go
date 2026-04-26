package gowebapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/felixge/httpsnoop"
	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs every HTTP request and
// response with full body, headers, and duration.
func logRequestResponse(logger golog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read and restore request body.
			var requestBodyMap map[string]interface{}
			if r.Body != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				if len(bodyBytes) > 0 {
					if err := json.Unmarshal(bodyBytes, &requestBodyMap); err != nil {
						requestBodyMap = map[string]interface{}{"raw": string(bodyBytes)}
					}
				}
			}

			response := responseCapture{statusCode: http.StatusOK}
			wrapped := httpsnoop.Wrap(w, httpsnoop.Hooks{
				Write: func(next httpsnoop.WriteFunc) httpsnoop.WriteFunc {
					return func(p []byte) (int, error) {
						response.body.Write(p)
						return next(p)
					}
				},
				WriteHeader: func(next httpsnoop.WriteHeaderFunc) httpsnoop.WriteHeaderFunc {
					return func(code int) {
						response.statusCode = code
						next(code)
					}
				},
			})

			start := time.Now()
			next.ServeHTTP(wrapped, r)
			duration := time.Since(start)

			var responseBodyMap map[string]interface{}
			responseBody := response.body.Bytes()
			if len(responseBody) > 0 {
				if err := json.Unmarshal(responseBody, &responseBodyMap); err != nil {
					responseBodyMap = map[string]interface{}{"raw": string(responseBody)}
				}
			}

			logger.Info(r.Context(), "http_request",
				golog.Field("request", map[string]interface{}{
					"body":    requestBodyMap,
					"headers": r.Header,
					"method":  r.Method,
					"path":    r.URL.Path,
					"query":   r.URL.Query(),
				}),
				golog.Field("response", map[string]interface{}{
					"body":     responseBodyMap,
					"duration": duration.Milliseconds(),
					"headers":  w.Header(),
					"status":   response.statusCode,
				}),
			)
		})
	}
}

type responseCapture struct {
	statusCode int
	body       bytes.Buffer
}
