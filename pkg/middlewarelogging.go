package gowebapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs every HTTP request and
// response with request details, response metadata, and duration.
func logRequestResponse(logger golog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBody any
			if r.Body != nil {
				bodyBytes, _ := io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				requestBody = parseLogBody(bodyBytes)
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var responseBody bytes.Buffer
			ww.Tee(&responseBody)

			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			logger.Info(r.Context(), "http_request",
				golog.Field("request", map[string]any{
					"body":    requestBody,
					"headers": r.Header,
					"query":   r.URL.Query(),
				}),
				golog.Field("response", map[string]any{
					"body":     parseLogBody(responseBody.Bytes()),
					"duration": duration.Milliseconds(),
					"headers":  ww.Header(),
					"status":   status,
				}),
			)
		})
	}
}

func parseLogBody(body []byte) any {
	if len(body) == 0 {
		return nil
	}

	var parsed any
	if err := json.Unmarshal(body, &parsed); err == nil {
		return parsed
	}

	return string(body)
}
