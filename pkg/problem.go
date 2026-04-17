package gowebapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Problem represents an RFC 7807 Problem Details object.
// See https://datatracker.ietf.org/doc/html/rfc7807
type Problem struct {
	// Type is a URI reference that identifies the problem type.
	Type string `json:"type"`
	// Title is a short, human-readable summary of the problem type.
	Title string `json:"title"`
	// Status is the HTTP status code for this occurrence.
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence.
	Instance string `json:"instance,omitempty"`
}

// HTTPError maps domain errors to HTTP Problem responses.
type HTTPError struct {
	Status int
	Title  string
	Detail string
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Title, e.Detail)
	}
	return e.Title
}

// WriteError maps an error to an HTTP Problem Details response.
// If the error is an *HTTPError, it uses the error's status and detail.
// Otherwise, it returns a generic 500 Internal Server Error.
func WriteError(w http.ResponseWriter, err error) {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		writeProblem(w, httpErr.Status, httpErr.Title, httpErr.Detail)
		return
	}

	writeProblem(w, http.StatusInternalServerError, "Internal Server Error", "")
}

// writeProblem writes a Problem Details JSON response.
func writeProblem(w http.ResponseWriter, status int, title, detail string) {
	p := Problem{
		Type:   "about:blank",
		Title:  title,
		Status: status,
		Detail: detail,
	}

	w.Header().Set("Content-Type", "application/problem+json; charset=utf-8")
	w.WriteHeader(status)

	body, err := json.Marshal(p)
	if err != nil {
		return
	}
	w.Write(body)
}

// --- Error constructors ---

// BadRequest creates a 400 Bad Request error.
func BadRequest(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusBadRequest, Title: "Bad Request", Detail: detail}
}

// Unauthorized creates a 401 Unauthorized error.
func Unauthorized(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusUnauthorized, Title: "Unauthorized", Detail: detail}
}

// Forbidden creates a 403 Forbidden error.
func Forbidden(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusForbidden, Title: "Forbidden", Detail: detail}
}

// NotFound creates a 404 Not Found error.
func NotFound(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusNotFound, Title: "Not Found", Detail: detail}
}

// Conflict creates a 409 Conflict error.
func Conflict(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusConflict, Title: "Conflict", Detail: detail}
}

// TooManyRequests creates a 429 Too Many Requests error.
func TooManyRequests(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusTooManyRequests, Title: "Too Many Requests", Detail: detail}
}

// InternalServerError creates a 500 Internal Server Error.
func InternalServerError(detail string) *HTTPError {
	return &HTTPError{Status: http.StatusInternalServerError, Title: "Internal Server Error", Detail: detail}
}
