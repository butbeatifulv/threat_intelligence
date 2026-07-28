package api

import (
	"encoding/json"
	"net/http"
)

// RegisterHealth attaches GET /health returning ok, service, and optional extra fields.
func RegisterHealth(mux *http.ServeMux, service string, extra map[string]any) {
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{"ok": true, "service": service}
		for k, v := range extra {
			body[k] = v
		}
		WriteJSON(w, http.StatusOK, body)
	})
}

// PostJSON registers a route that decodes JSON body into a map and calls fn.
func PostJSON(mux *http.ServeMux, pattern string, fn func(*http.Request, map[string]any) (any, int)) {
	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&body)
		}
		if body == nil {
			body = map[string]any{}
		}
		res, code := fn(r, body)
		WriteJSON(w, code, res)
	})
}
