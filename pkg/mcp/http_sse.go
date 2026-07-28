package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// WantsSSE reports whether the client prefers SSE responses.
func WantsSSE(r *http.Request, preferSSE bool) bool {
	if preferSSE {
		return true
	}
	accept := strings.ToLower(r.Header.Get("Accept"))
	if strings.Contains(accept, "text/event-stream") {
		return true
	}
	return containsMIME(accept, "text/event-stream")
}

func containsMIME(accept, mime string) bool {
	for _, part := range splitAccept(accept) {
		if part == mime || part == "*/*" {
			return true
		}
	}
	return false
}

func splitAccept(h string) []string {
	var out []string
	for _, p := range splitComma(h) {
		if i := indexByte(p, ';'); i >= 0 {
			p = trimSpace(p[:i])
		}
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}

// WriteSSEMessages streams JSON-RPC messages as text/event-stream.
func WriteSSEMessages(w http.ResponseWriter, messages []Message) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming not supported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	for _, msg := range messages {
		b, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: message\ndata: %s\n\n", b); err != nil {
			return err
		}
		flusher.Flush()
	}
	return nil
}
