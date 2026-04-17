package gowebapp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *HTTPError
		expected string
	}{
		{
			name:     "with detail",
			err:      BadRequest("invalid email format"),
			expected: "Bad Request: invalid email format",
		},
		{
			name: "without detail",
			err: &HTTPError{
				Status: http.StatusBadRequest,
				Title:  "Bad Request",
			},
			expected: "Bad Request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestWriteError_HTTPError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedTitle  string
		expectedDetail string
	}{
		{
			name:           "bad request",
			err:            BadRequest("missing field 'name'"),
			expectedStatus: 400,
			expectedTitle:  "Bad Request",
			expectedDetail: "missing field 'name'",
		},
		{
			name:           "unauthorized",
			err:            Unauthorized("invalid token"),
			expectedStatus: 401,
			expectedTitle:  "Unauthorized",
			expectedDetail: "invalid token",
		},
		{
			name:           "forbidden",
			err:            Forbidden("insufficient permissions"),
			expectedStatus: 403,
			expectedTitle:  "Forbidden",
			expectedDetail: "insufficient permissions",
		},
		{
			name:           "not found",
			err:            NotFound("user not found"),
			expectedStatus: 404,
			expectedTitle:  "Not Found",
			expectedDetail: "user not found",
		},
		{
			name:           "conflict",
			err:            Conflict("resource already exists"),
			expectedStatus: 409,
			expectedTitle:  "Conflict",
			expectedDetail: "resource already exists",
		},
		{
			name:           "too many requests",
			err:            TooManyRequests("rate limit exceeded"),
			expectedStatus: 429,
			expectedTitle:  "Too Many Requests",
			expectedDetail: "rate limit exceeded",
		},
		{
			name:           "internal server error",
			err:            InternalServerError("database connection failed"),
			expectedStatus: 500,
			expectedTitle:  "Internal Server Error",
			expectedDetail: "database connection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			WriteError(rr, tt.err)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "application/problem+json; charset=utf-8", rr.Header().Get("Content-Type"))

			var problem Problem
			err := json.Unmarshal(rr.Body.Bytes(), &problem)
			require.NoError(t, err)

			assert.Equal(t, "about:blank", problem.Type)
			assert.Equal(t, tt.expectedTitle, problem.Title)
			assert.Equal(t, tt.expectedStatus, problem.Status)
			assert.Equal(t, tt.expectedDetail, problem.Detail)
		})
	}
}

func TestWriteError_GenericError(t *testing.T) {
	rr := httptest.NewRecorder()
	WriteError(rr, errors.New("something went wrong"))

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	var problem Problem
	err := json.Unmarshal(rr.Body.Bytes(), &problem)
	require.NoError(t, err)

	assert.Equal(t, "Internal Server Error", problem.Title)
	assert.Equal(t, 500, problem.Status)
	assert.Empty(t, problem.Detail)
}

func TestErrorConstructors(t *testing.T) {
	constructors := []struct {
		name   string
		fn     func(string) *HTTPError
		status int
	}{
		{"BadRequest", BadRequest, 400},
		{"Unauthorized", Unauthorized, 401},
		{"Forbidden", Forbidden, 403},
		{"NotFound", NotFound, 404},
		{"Conflict", Conflict, 409},
		{"TooManyRequests", TooManyRequests, 429},
		{"InternalServerError", InternalServerError, 500},
	}

	for _, tc := range constructors {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.fn("test detail")
			assert.Equal(t, tc.status, err.Status)
			assert.NotEmpty(t, err.Title)
			assert.Equal(t, "test detail", err.Detail)
			var e error = err
			assert.NotNil(t, e)
		})
	}
}
