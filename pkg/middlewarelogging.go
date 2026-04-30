package gowebapp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	golog "github.com/marcosstupnicki/go-log"
)

// logRequestResponse creates middleware that logs every HTTP request and
// response with request details, response metadata, and duration.
func logRequestResponse(logger golog.Logger, config HTTPLoggingConfig) func(http.Handler) http.Handler {
	config = normalizeHTTPLoggingConfig(config)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			var requestBody any
			if config.IncludeRequestBody && r.Body != nil {
				readBytes, logBytes, truncated, err := readBodyPrefix(r.Body, config.MaxBodyBytes)
				if err == nil {
					r.Body = prependReadCloser(readBytes, r.Body)
					requestBody = parseLogBody(logBytes, truncated)
				}
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			var responseBody *limitedBuffer
			if config.IncludeResponseBody {
				responseBody = newLimitedBuffer(config.MaxBodyBytes)
				ww.Tee(responseBody)
			}

			next.ServeHTTP(ww, r)
			duration := time.Since(start)

			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			request := map[string]any{}
			if config.IncludeRequestBody {
				request["body"] = requestBody
			}
			if config.IncludeRequestHeaders {
				request["headers"] = redactHeaders(r.Header, config.RedactedHeaders)
			}
			if config.IncludeRequestQuery {
				request["query"] = redactValues(r.URL.Query(), config.RedactedQueryParams)
			}

			response := map[string]any{
				"duration": duration.Milliseconds(),
				"status":   status,
			}
			if config.IncludeResponseBody && responseBody != nil {
				response["body"] = parseLogBody(responseBody.Bytes(), responseBody.Truncated())
			}
			if config.IncludeResponseHeaders {
				response["headers"] = redactHeaders(ww.Header(), config.RedactedHeaders)
			}

			logger.Info(r.Context(), "http_request",
				golog.Field("request", request),
				golog.Field("response", response),
			)
		})
	}
}

func parseLogBody(body []byte, truncated bool) any {
	if len(body) == 0 {
		return nil
	}

	var value any
	var parsed any
	if err := json.Unmarshal(body, &parsed); err == nil {
		value = parsed
	} else {
		value = string(body)
	}

	if !truncated {
		return value
	}

	return map[string]any{
		"truncated": true,
		"value":     value,
	}
}

func readBodyPrefix(body io.Reader, limit int64) ([]byte, []byte, bool, error) {
	if limit <= 0 {
		return nil, nil, false, nil
	}
	readBytes, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, nil, false, err
	}
	if int64(len(readBytes)) > limit {
		return readBytes, readBytes[:limit], true, nil
	}
	return readBytes, readBytes, false, nil
}

type prependCloser struct {
	reader io.Reader
	closer io.Closer
}

func prependReadCloser(prefix []byte, body io.ReadCloser) io.ReadCloser {
	return &prependCloser{
		reader: io.MultiReader(bytes.NewReader(prefix), body),
		closer: body,
	}
}

func (p *prependCloser) Read(b []byte) (int, error) {
	return p.reader.Read(b)
}

func (p *prependCloser) Close() error {
	return p.closer.Close()
}

type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int64
	truncated bool
}

func newLimitedBuffer(limit int64) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		b.truncated = b.truncated || len(p) > 0
		return len(p), nil
	}

	remaining := b.limit - int64(b.buf.Len())
	if remaining <= 0 {
		b.truncated = b.truncated || len(p) > 0
		return len(p), nil
	}
	if int64(len(p)) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}
	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) Bytes() []byte {
	return b.buf.Bytes()
}

func (b *limitedBuffer) Truncated() bool {
	return b.truncated
}

func redactHeaders(headers http.Header, redacted []string) map[string][]string {
	if headers == nil {
		return nil
	}
	redactedSet := stringSet(redacted)
	out := make(map[string][]string, len(headers))
	for key, values := range headers {
		if _, ok := redactedSet[strings.ToLower(key)]; ok {
			out[key] = []string{"[REDACTED]"}
			continue
		}
		out[key] = cloneStringSlice(values)
	}
	return out
}

func redactValues(values url.Values, redacted []string) map[string][]string {
	if values == nil {
		return nil
	}
	redactedSet := stringSet(redacted)
	out := make(map[string][]string, len(values))
	for key, value := range values {
		if _, ok := redactedSet[strings.ToLower(key)]; ok {
			out[key] = []string{"[REDACTED]"}
			continue
		}
		out[key] = cloneStringSlice(value)
	}
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[strings.ToLower(value)] = struct{}{}
	}
	return out
}
