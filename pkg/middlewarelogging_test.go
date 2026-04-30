package gowebapp

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogRequestResponseShapesRequestAndResponseBodies(t *testing.T) {
	output := captureStderr(t, func() {
		webapp := mustNew(t, "test", "8080", WithHTTPLogging(VerboseHTTPLoggingConfig()))
		webapp.Post("/hooks/{sourceID}", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"accepted":true,"provider":"xml"}`))
		})

		req := httptest.NewRequest(http.MethodPost, "/hooks/source-xml?provider=custom", strings.NewReader(`<event id="xml_local">ok</event>`))
		req.Header.Set("Content-Type", "application/xml")
		rr := httptest.NewRecorder()

		webapp.Router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
	})

	entry := logEntryByMessage(t, output, "http_request")

	require.Equal(t, http.MethodPost, entry["method"])
	require.Equal(t, "/hooks/source-xml", entry["path"])

	request, ok := entry["request"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, `<event id="xml_local">ok</event>`, request["body"])
	require.NotContains(t, request, "method")
	require.NotContains(t, request, "path")

	response, ok := entry["response"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(http.StatusAccepted), response["status"])
	require.Contains(t, response, "duration")
	require.NotContains(t, response, "bytes")
	require.NotContains(t, response, "duration_ms")

	body, ok := response["body"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, body["accepted"])
	require.Equal(t, "xml", body["provider"])
}

func TestLogRequestResponseDefaultOmitsSensitiveDetails(t *testing.T) {
	output := captureStderr(t, func() {
		webapp := mustNew(t, "test", "8080")
		webapp.Post("/hooks", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"secret":"response"}`))
		})

		req := httptest.NewRequest(http.MethodPost, "/hooks?token=secret-token", strings.NewReader(`{"secret":"request"}`))
		req.Header.Set("Authorization", "Bearer secret-token")
		rr := httptest.NewRecorder()

		webapp.Router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusAccepted, rr.Code)
	})

	entry := logEntryByMessage(t, output, "http_request")
	request, ok := entry["request"].(map[string]any)
	require.True(t, ok)
	require.Empty(t, request)

	response, ok := entry["response"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, float64(http.StatusAccepted), response["status"])
	require.Contains(t, response, "duration")
	require.NotContains(t, response, "body")
	require.NotContains(t, response, "headers")
}

func TestLogRequestResponseRedactsHeadersAndQuery(t *testing.T) {
	output := captureStderr(t, func() {
		webapp := mustNew(t, "test", "8080", WithHTTPLogging(HTTPLoggingConfig{
			Enabled:                true,
			IncludeRequestHeaders:  true,
			IncludeRequestQuery:    true,
			IncludeResponseHeaders: true,
		}))
		webapp.Get("/hooks", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Set-Cookie", "session=secret")
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/hooks?token=secret-token&provider=stripe", nil)
		req.Header.Set("Authorization", "Bearer secret-token")
		req.Header.Set("X-Trace", "visible")
		rr := httptest.NewRecorder()

		webapp.Router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	entry := logEntryByMessage(t, output, "http_request")
	request := entry["request"].(map[string]any)
	headers := request["headers"].(map[string]any)
	query := request["query"].(map[string]any)
	require.Equal(t, []any{"[REDACTED]"}, headers["Authorization"])
	require.Equal(t, []any{"visible"}, headers["X-Trace"])
	require.Equal(t, []any{"[REDACTED]"}, query["token"])
	require.Equal(t, []any{"stripe"}, query["provider"])

	response := entry["response"].(map[string]any)
	responseHeaders := response["headers"].(map[string]any)
	require.Equal(t, []any{"[REDACTED]"}, responseHeaders["Set-Cookie"])
}

func TestLogRequestResponseTruncatesRequestBodyWithoutChangingHandlerBody(t *testing.T) {
	const body = "abcdef"

	output := captureStderr(t, func() {
		webapp := mustNew(t, "test", "8080", WithHTTPLogging(HTTPLoggingConfig{
			Enabled:            true,
			IncludeRequestBody: true,
			MaxBodyBytes:       3,
		}))
		webapp.Post("/hooks", func(w http.ResponseWriter, r *http.Request) {
			got, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.Equal(t, body, string(got))
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodPost, "/hooks", strings.NewReader(body))
		rr := httptest.NewRecorder()

		webapp.Router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNoContent, rr.Code)
	})

	entry := logEntryByMessage(t, output, "http_request")
	request := entry["request"].(map[string]any)
	loggedBody := request["body"].(map[string]any)
	require.Equal(t, true, loggedBody["truncated"])
	require.Equal(t, "abc", loggedBody["value"])
}

func logEntryByMessage(t *testing.T, output string, msg string) map[string]any {
	t.Helper()

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}

		var entry map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &entry))
		if entry["msg"] == msg {
			return entry
		}
	}

	t.Fatalf("log entry with msg %q not found in output:\n%s", msg, output)
	return nil
}
