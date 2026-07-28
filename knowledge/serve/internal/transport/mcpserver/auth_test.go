package mcpserver

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/auth/keycloak"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
)

func TestServer_toolsCall_requiresAuth(t *testing.T) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	issuer := "https://keycloak.example/realms/veil"
	aud := "veil-api"
	v := keycloak.NewStaticVerifier(issuer, aud, "veil-api", &key.PublicKey)
	stack := auth.NewStack(v, auth.Config{
		Enabled:        true,
		RBACEnabled:    true,
		RoleReader:     "veil-reader",
		MCPAccessToken: "",
	})

	uc := usecase.NewReadUsecase(&mockExec{})
	srv := NewServer(uc, nil, nil, nil, stack, slog.Default())

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	go func() { _ = srv.Run(context.Background(), stdinR, stdoutW) }()

	write := func(id int, method string, params any) {
		b, _ := json.Marshal(params)
		rw := newFramedRW(strings.NewReader(""), stdinW)
		_ = rw.WriteJSON(context.Background(), rpcMessage{
			JSONRPC: "2.0", ID: id, Method: method, Params: b,
		})
	}
	read := func() rpcMessage {
		rw := newFramedRW(stdoutR, io.Discard)
		payload, err := rw.Read(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		var msg rpcMessage
		_ = json.Unmarshal(payload, &msg)
		return msg
	}

	write(1, "tools/call", map[string]any{
		"name": "ti_health", "arguments": map[string]any{},
	})
	resp := read()
	if resp.Error == nil || resp.Error.Code != codeAuthError {
		t.Fatalf("expected auth error, got %+v", resp.Error)
	}

	tok, _ := keycloak.SignTestToken(key, issuer, aud, "u1", []string{"veil-reader"}, time.Hour)
	stack.Config.MCPAccessToken = tok
	srv2 := NewServer(uc, nil, nil, nil, stack, slog.Default())
	stdinR2, stdinW2 := io.Pipe()
	stdoutR2, stdoutW2 := io.Pipe()
	go func() { _ = srv2.Run(context.Background(), stdinR2, stdoutW2) }()
	write2 := func(id int, method string, params any) {
		b, _ := json.Marshal(params)
		rw := newFramedRW(strings.NewReader(""), stdinW2)
		_ = rw.WriteJSON(context.Background(), rpcMessage{
			JSONRPC: "2.0", ID: id, Method: method, Params: b,
		})
	}
	read2 := func() rpcMessage {
		rw := newFramedRW(stdoutR2, io.Discard)
		payload, _ := rw.Read(context.Background())
		var msg rpcMessage
		_ = json.Unmarshal(payload, &msg)
		return msg
	}
	write2(2, "tools/call", map[string]any{
		"name": "ti_health", "arguments": map[string]any{},
	})
	resp2 := read2()
	if resp2.Error != nil {
		t.Fatalf("unexpected error: %+v", resp2.Error)
	}
}
