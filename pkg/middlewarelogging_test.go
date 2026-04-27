package gowebapp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLogRequestResponseShapesRequestAndResponseBodies(t *testing.T) {
	output := captureStderr(t, func() {
		webapp := mustNew(t, "test", "8080")
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
