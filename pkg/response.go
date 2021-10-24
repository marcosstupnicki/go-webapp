package gowebapp

import (
	"encoding/json"
	"net/http"
)

// RespondWithJSON sends the response "payload" as JSON format to the client with code as status code.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	if code == 0 {
		return ErrStatusCode0Invalid
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	if !isBodyAllowed(code) {
		return nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err := w.Write(body); err != nil {
		return err
	}

	return nil
}

// isBodyAllowed reports whether a given response status code permits a body.
// See RFC 7230, section 3.3.
func isBodyAllowed(status int) bool {
	if (status >= 100 && status <= 199) || status == 204 || status == 304 {
		return false
	}

	return true
}

// RespondWithError return error message as JSON format. Eg: {"message": "Access denied for user 'root'"}
func RespondWithError(w http.ResponseWriter, code int, msg string) error {
	return RespondWithJSON(w, code, map[string]string{"message": msg})
}
