package api

import (
	"encoding/json"
	"net/http"
)

// WriteJSON sets Content-Type application/json, status, and encodes v.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// SanitizeError maps internal errors to generic messages when prod mode is on.
func SanitizeError(status int, msg string) string {
	if !prodMode.Load() {
		return msg
	}
	switch {
	case status == http.StatusNotFound:
		return "not found"
	case status >= 500:
		return "internal error"
	default:
		return "bad request"
	}
}

// WriteError writes {"error": msg} with optional prod sanitization.
func WriteError(w http.ResponseWriter, status int, err error) {
	msg := err.Error()
	msg = SanitizeError(status, msg)
	WriteJSON(w, status, map[string]any{"error": msg})
}
