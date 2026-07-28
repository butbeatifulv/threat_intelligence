package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWantsSSE_preferSSE(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	if !WantsSSE(req, true) {
		t.Fatal("expected true when preferSSE")
	}
}

func TestWantsSSE_acceptHeader(t *testing.T) {
	tests := []struct {
		name   string
		accept string
		want   bool
	}{
		{"substring match", "application/json, text/event-stream", true},
		{"mime list", "application/json;q=0.9, text/event-stream;q=1", true},
		{"wildcard", "*/*", true},
		{"json only", "application/json", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}
			if got := WantsSSE(req, false); got != tt.want {
				t.Fatalf("WantsSSE() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsMIME(t *testing.T) {
	tests := []struct {
		accept string
		mime   string
		want   bool
	}{
		{"text/event-stream", "text/event-stream", true},
		{"application/json,text/event-stream", "text/event-stream", true},
		{"application/json;q=0.8,*/*", "text/event-stream", true},
		{"application/json", "text/event-stream", false},
	}
	for _, tt := range tests {
		t.Run(tt.accept, func(t *testing.T) {
			if got := containsMIME(tt.accept, tt.mime); got != tt.want {
				t.Fatalf("containsMIME(%q, %q) = %v, want %v", tt.accept, tt.mime, got, tt.want)
			}
		})
	}
}
