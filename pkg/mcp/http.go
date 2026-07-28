package mcp

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

// HTTPConfig configures the streamable HTTP MCP handler.
type HTTPConfig struct {
	Path       string
	PreferSSE  bool
	Service    string
	HealthExtra map[string]any
	Logger     *slog.Logger
}

// HTTPHandler serves Streamable HTTP MCP (POST JSON or SSE).
func HTTPHandler(proc Processor, cfg HTTPConfig) http.Handler {
	mux := http.NewServeMux()
	path := cfg.Path
	if path == "" {
		path = "/mcp"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		body := map[string]any{
			"ok":        true,
			"service":   cfg.Service,
			"transport": "streamable-http",
		}
		for k, v := range cfg.HealthExtra {
			body[k] = v
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	})

	mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("POST "+path, func(w http.ResponseWriter, r *http.Request) {
		serveMCPPost(w, r, proc, cfg)
	})

	return mux
}

func serveMCPPost(w http.ResponseWriter, r *http.Request, proc Processor, cfg HTTPConfig) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 4<<20))
	if err != nil {
		writeHTTPRPCError(w, http.StatusBadRequest, CodeParseError, "read body failed")
		return
	}
	if len(body) == 0 {
		writeHTTPRPCError(w, http.StatusBadRequest, CodeInvalidRequest, "empty body")
		return
	}

	msgs, err := ParseInboundMessages(body)
	if err != nil {
		writeHTTPRPCError(w, http.StatusBadRequest, CodeParseError, err.Error())
		return
	}

	var requests, notifications []Message
	for _, m := range msgs {
		if m.Method != "" && m.ID != nil {
			requests = append(requests, m)
		} else if m.Method != "" && m.ID == nil {
			notifications = append(notifications, m)
		}
	}

	if len(requests) == 0 {
		for _, n := range notifications {
			_, _, _ = proc.ProcessMessage(r.Context(), n, true)
		}
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if len(msgs) > 1 {
		writeHTTPRPCError(w, http.StatusBadRequest, CodeInvalidRequest, "batch not supported")
		return
	}

	var out []Message
	for _, req := range requests {
		resp, isNotification, perr := proc.ProcessMessage(r.Context(), req, true)
		if perr != nil {
			writeHTTPRPCError(w, http.StatusInternalServerError, CodeInternal, perr.Error())
			return
		}
		if isNotification || resp == nil {
			continue
		}
		out = append(out, *resp)
	}

	if len(out) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	if WantsSSE(r, cfg.PreferSSE) {
		if err := WriteSSEMessages(w, out); err != nil && cfg.Logger != nil {
			cfg.Logger.Error("mcp sse write failed", "err", err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(out[0])
}

// ParseInboundMessages decodes a single object or JSON array batch.
func ParseInboundMessages(body []byte) ([]Message, error) {
	body = trimSpaceBytes(body)
	if len(body) > 0 && body[0] == '[' {
		var batch []Message
		if err := json.Unmarshal(body, &batch); err != nil {
			return nil, err
		}
		return batch, nil
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return []Message{msg}, nil
}

func trimSpaceBytes(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\n' || b[0] == '\r' || b[0] == '\t') {
		b = b[1:]
	}
	for len(b) > 0 && (b[len(b)-1] == ' ' || b[len(b)-1] == '\n' || b[len(b)-1] == '\r' || b[len(b)-1] == '\t') {
		b = b[:len(b)-1]
	}
	return b
}

func writeHTTPRPCError(w http.ResponseWriter, status, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Message{
		JSONRPC: "2.0",
		Error:   &RPCError{Code: code, Message: msg},
	})
}
