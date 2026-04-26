package gowebapp

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs every HTTP request and
// response with status, byte count, and duration.
func logRequestResponse(logger golog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			logger.Info(r.Context(), "http_request",
				golog.Field("status", ww.Status()),
				golog.Field("bytes", ww.BytesWritten()),
				golog.Field("duration_ms", duration.Milliseconds()),
			)
		})
	}
}
