package gowebapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs HTTP requests and
// responses. It respects body size limits and skip flags from config.
func logRequestResponse(logger golog.Logger, cfg webAppConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip logging for health check endpoint
			if r.URL.Path == cfg.healthPath {
				next.ServeHTTP(w, r)
				return
			}

			// Capture request body (with size limit)
			var requestBodyMap map[string]interface{}
			if !cfg.skipRequestBody && r.Body != nil {
				requestBodyMap = readAndRestoreBody(r, cfg.maxBodyLogSize)
			}

			// Wrap response writer to capture status and body
			lrw := newLoggingResponseWriter(w, cfg.maxBodyLogSize, cfg.skipResponseBody)

			startTime := time.Now()
			next.ServeHTTP(lrw, r)
			duration := time.Since(startTime)

			// Build log entry
			requestLog := map[string]interface{}{
				"method":  r.Method,
				"path":    r.URL.Path,
				"query":   r.URL.Query(),
				"headers": r.Header,
			}
			if !cfg.skipRequestBody {
				requestLog["body"] = requestBodyMap
			}

			responseLog := map[string]interface{}{
				"status":   lrw.statusCode,
				"headers":  lrw.Header(),
				"duration": duration.Milliseconds(),
			}
			if !cfg.skipResponseBody {
				responseLog["body"] = lrw.GetBodyMap()
			}

			// Log using the request context so Enriched fields (request_id, etc.) are included
			logger.Info(r.Context(), "http_request",
				golog.Field("request", requestLog),
				golog.Field("response", responseLog),
			)
		})
	}
}

// readAndRestoreBody reads up to maxSize bytes from the request body,
// restores the body for downstream handlers, and returns the decoded map.
func readAndRestoreBody(r *http.Request, maxSize int) map[string]interface{} {
	if r.Body == nil {
		return nil
	}

	limitReader := io.LimitReader(r.Body, int64(maxSize+1))
	bodyBytes, err := io.ReadAll(limitReader)
	if err != nil {
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		return nil
	}

	truncated := len(bodyBytes) > maxSize
	if truncated {
		remaining, _ := io.ReadAll(r.Body)
		fullBody := append(bodyBytes, remaining...)
		r.Body = io.NopCloser(bytes.NewBuffer(fullBody))
		bodyBytes = bodyBytes[:maxSize]
	} else {
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	if len(bodyBytes) == 0 {
		return nil
	}

	var bodyMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		raw := string(bodyBytes)
		if truncated {
			raw += "[truncated]"
		}
		return map[string]interface{}{"raw": raw}
	}

	if truncated {
		bodyMap["_truncated"] = true
	}

	return bodyMap
}

// LoggingResponseWriter wraps http.ResponseWriter to capture status
// code and response body.
type LoggingResponseWriter struct {
	http.ResponseWriter
	statusCode    int
	body          bytes.Buffer
	maxSize       int
	skipBody      bool
	bodyTruncated bool
	bytesWritten  int
}

func newLoggingResponseWriter(w http.ResponseWriter, maxSize int, skipBody bool) *LoggingResponseWriter {
	return &LoggingResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		maxSize:        maxSize,
		skipBody:       skipBody,
	}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	if !lrw.skipBody && !lrw.bodyTruncated {
		remaining := lrw.maxSize - lrw.bytesWritten
		if remaining > 0 {
			toCapture := b
			if len(toCapture) > remaining {
				toCapture = toCapture[:remaining]
				lrw.bodyTruncated = true
			}
			lrw.body.Write(toCapture)
		} else {
			lrw.bodyTruncated = true
		}
		lrw.bytesWritten += len(b)
	}
	return lrw.ResponseWriter.Write(b)
}

func (lrw *LoggingResponseWriter) GetBody() []byte {
	return lrw.body.Bytes()
}

func (lrw *LoggingResponseWriter) GetBodyMap() map[string]interface{} {
	responseBody := lrw.GetBody()
	if len(responseBody) == 0 {
		return nil
	}

	var responseBodyMap map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseBodyMap); err != nil {
		raw := string(responseBody)
		if lrw.bodyTruncated {
			raw += "[truncated]"
		}
		return map[string]interface{}{"raw": raw}
	}

	if lrw.bodyTruncated {
		responseBodyMap["_truncated"] = true
	}

	return responseBodyMap
}
