package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func TestToRPCError_andRPCErrorString(t *testing.T) {
	orig := &RPCError{Code: CodeInvalidParams, Message: "bad"}
	if got := ToRPCError(orig); got != orig {
		t.Fatalf("got %+v", got)
	}
	if got := ToRPCError(auth.ErrForbidden); got.Code != CodeAuthError || got.Message != "forbidden" {
		t.Fatalf("forbidden: %+v", got)
	}
	if got := ToRPCError(auth.ErrUnauthorized); got.Message != "unauthorized" {
		t.Fatalf("unauthorized: %+v", got)
	}
	if got := ToRPCError(errors.New("boom")); got.Code != CodeInternal || got.Message != "boom" {
		t.Fatalf("internal: %+v", got)
	}
	if (&RPCError{Message: "x"}).Error() != "x" {
		t.Fatal("Error() string")
	}
}

func TestParseInboundMessages_whitespaceOnly(t *testing.T) {
	if _, err := ParseInboundMessages([]byte("  \t  ")); err == nil {
		t.Fatal("expected error")
	}
}

func TestParseInboundMessages_trimSpace(t *testing.T) {
	body := []byte("\n\t {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"ping\"} \r\n\t")
	msgs, err := ParseInboundMessages(body)
	if err != nil || len(msgs) != 1 || msgs[0].Method != "ping" {
		t.Fatalf("msgs=%v err=%v", msgs, err)
	}
}

func TestNegotiateProtocol_emptyVersion(t *testing.T) {
	params, _ := json.Marshal(initializeParams{ProtocolVersion: ""})
	if got := NegotiateProtocol(params); got != DefaultProtocol {
		t.Fatalf("got %q", got)
	}
}

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (f *flushRecorder) Flush() { f.flushed = true }

func TestWriteSSEMessages_ok(t *testing.T) {
	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	msgs := []Message{{JSONRPC: "2.0", ID: 1, Result: map[string]any{"ok": true}}}
	if err := WriteSSEMessages(rec, msgs); err != nil {
		t.Fatal(err)
	}
	if !rec.flushed || !bytes.Contains(rec.Body.Bytes(), []byte("event: message")) {
		t.Fatalf("body=%q flushed=%v", rec.Body.Bytes(), rec.flushed)
	}
}

type noFlushResponseWriter struct {
	header http.Header
	code   int
	body   bytes.Buffer
}

func (w *noFlushResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}
func (w *noFlushResponseWriter) Write(b []byte) (int, error) { return w.body.Write(b) }
func (w *noFlushResponseWriter) WriteHeader(status int)      { w.code = status }

func TestWriteSSEMessages_noFlusher(t *testing.T) {
	err := WriteSSEMessages(&noFlushResponseWriter{}, nil)
	if err == nil || err.Error() != "streaming not supported" {
		t.Fatalf("got %v", err)
	}
}

func TestTrimSpace_viaSplitAccept(t *testing.T) {
	if got := splitAccept(" application/json ;q=0.9\t"); len(got) != 1 || got[0] != "application/json" {
		t.Fatalf("got %v", got)
	}
	if indexByte("abc", 'z') != -1 {
		t.Fatal("expected -1")
	}
}

func TestHTTP_healthAndDefaults(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{
		Path:        "",
		Service:     "svc",
		HealthExtra: map[string]any{"extra": true},
	})
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["service"] != "svc" || body["extra"] != true {
		t.Fatalf("body %v", body)
	}
}

type errProc struct{ err error }

func (e errProc) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	return nil, false, e.err
}

func TestHTTP_post_readBodyError(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "/mcp"})
	req := httptest.NewRequest(http.MethodPost, "/mcp", errReadCloser{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

type errReadCloser struct{}

func (errReadCloser) Read([]byte) (int, error) { return 0, errors.New("read failed") }
func (errReadCloser) Close() error             { return nil }

func TestHTTP_post_successJSON(t *testing.T) {
	proc := jsonReplyProc{}
	h := HTTPHandler(proc, HTTPConfig{Path: "/mcp"})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	req.Header.Set("Accept", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
}

func TestHTTP_post_ssePrefer(t *testing.T) {
	proc := jsonReplyProc{}
	h := HTTPHandler(proc, HTTPConfig{Path: "/mcp", PreferSSE: true})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestHTTP_post_processorError(t *testing.T) {
	h := HTTPHandler(errProc{err: errors.New("proc fail")}, HTTPConfig{Path: "/mcp"})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status %d", rec.Code)
	}
}

type jsonReplyProc struct{}

func (jsonReplyProc) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	return &Message{JSONRPC: "2.0", ID: msg.ID, Result: map[string]any{"ok": true}}, false, nil
}

type emptyRespProc struct{}

func (emptyRespProc) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	return nil, false, nil
}

func TestHTTP_post_acceptedNoBody(t *testing.T) {
	h := HTTPHandler(emptyRespProc{}, HTTPConfig{Path: "/mcp"})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestAuthorizeToolCall_unknownPermission(t *testing.T) {
	stack := testAuthStack(true, "veil-engage-runner")
	sub := &auth.Subject{Sub: "u1", Roles: []string{"veil-engage-runner"}}
	ctx := auth.WithSubject(context.Background(), sub)
	_, err := AuthorizeToolCall(ctx, stack, "unknown:perm", nil)
	var re *RPCError
	if !errors.As(err, &re) || re.Message != "unauthorized" {
		t.Fatalf("got %v", err)
	}
}

func TestAuthorizeToolCall_fallbackUnauthorized(t *testing.T) {
	stack := testAuthStack(true, "veil-engage-runner")
	fallback := func(ctx context.Context, s *auth.Stack, _ string) (context.Context, error) {
		return ctx, auth.ErrUnauthorized
	}
	_, err := AuthorizeToolCall(context.Background(), stack, auth.PermEngageToolRun, fallback)
	var re *RPCError
	if !errors.As(err, &re) || re.Message != "unauthorized" {
		t.Fatalf("got %v", err)
	}
}

func TestFramedRW_invalidFrame(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("Content-Length: 0\r\n\r\n")
	rw := NewFramedRW(&buf, &buf)
	_, err := rw.Read(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFramedRW_skipsUnknownHeaders(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("X-Ignore: yes\r\n")
	buf.WriteString("badline\r\n")
	buf.WriteString("Content-Length: 2\r\n\r\n")
	buf.WriteString("{}")
	rw := NewFramedRW(&buf, &buf)
	raw, err := rw.Read(context.Background())
	if err != nil || string(raw) != "{}" {
		t.Fatalf("raw=%q err=%v", raw, err)
	}
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, io.ErrShortWrite }

func TestFramedRW_marshalError(t *testing.T) {
	rw := NewFramedRW(&bytes.Buffer{}, &bytes.Buffer{})
	err := rw.WriteJSON(context.Background(), map[string]any{"bad": make(chan int)})
	if err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestFramedRW_writeError(t *testing.T) {
	rw := NewFramedRW(&bytes.Buffer{}, failWriter{})
	if err := rw.WriteJSON(context.Background(), Message{JSONRPC: "2.0"}); err == nil {
		t.Fatal("expected write error")
	}
}

func TestRunStdio_emptyMethodSkipped(t *testing.T) {
	var in bytes.Buffer
	msg := Message{JSONRPC: "2.0", ID: 1}
	b, _ := json.Marshal(msg)
	writeFramed(&in, b)
	err := RunStdio(context.Background(), mockProcessor{}, &in, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestRunStdio_writeFails(t *testing.T) {
	var in bytes.Buffer
	req := Message{JSONRPC: "2.0", ID: 1, Method: "ping"}
	b, _ := json.Marshal(req)
	writeFramed(&in, b)
	err := RunStdio(context.Background(), mockProcessor{}, &in, failWriter{})
	if err == nil {
		t.Fatal("expected write error")
	}
}

type notifyStdioProc struct{}

func (notifyStdioProc) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	if msg.Method == "notify" {
		return nil, true, nil
	}
	return &Message{JSONRPC: "2.0", ID: msg.ID, Result: true}, false, nil
}

func TestRunStdio_notificationNoResponse(t *testing.T) {
	var in bytes.Buffer
	req := Message{JSONRPC: "2.0", Method: "notify"}
	b, _ := json.Marshal(req)
	writeFramed(&in, b)
	err := RunStdio(context.Background(), notifyStdioProc{}, &in, &bytes.Buffer{})
	if err != nil {
		t.Fatal(err)
	}
}

type httpNotifyProc struct{}

func (httpNotifyProc) ProcessMessage(ctx context.Context, msg Message, httpTransport bool) (*Message, bool, error) {
	return nil, true, nil
}

func TestHTTP_post_notificationResponseSkipped(t *testing.T) {
	h := HTTPHandler(httpNotifyProc{}, HTTPConfig{Path: "/mcp"})
	body := `{"jsonrpc":"2.0","id":1,"method":"tools/call"}`
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status %d", rec.Code)
	}
}

type failFlushWriter struct {
	noFlushResponseWriter
	fail bool
}

func (w *failFlushWriter) Write(b []byte) (int, error) {
	if w.fail {
		return 0, errors.New("write failed")
	}
	return w.noFlushResponseWriter.Write(b)
}
func (w *failFlushWriter) Flush() {}

func TestWriteSSEMessages_marshalError(t *testing.T) {
	w := &failFlushWriter{}
	msgs := []Message{{JSONRPC: "2.0", Result: map[string]any{"bad": make(chan int)}}}
	if err := WriteSSEMessages(w, msgs); err == nil {
		t.Fatal("expected marshal error")
	}
}

func TestWriteSSEMessages_writeError(t *testing.T) {
	w := &failFlushWriter{fail: true}
	msgs := []Message{{JSONRPC: "2.0", ID: 1, Result: true}}
	if err := WriteSSEMessages(w, msgs); err == nil {
		t.Fatal("expected write error")
	}
}

func TestRunStdio_readError(t *testing.T) {
	var in bytes.Buffer
	in.WriteString("Content-Length: 5\r\n\r\n")
	in.WriteString("{\"x") // truncated body
	err := RunStdio(context.Background(), mockProcessor{}, &in, &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected read error")
	}
}

func TestHTTP_post_parseError(t *testing.T) {
	h := HTTPHandler(stubProc{}, HTTPConfig{Path: "/mcp"})
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString("{bad"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status %d", rec.Code)
	}
}

func TestServeMCPPost_sseWriteLogsError(t *testing.T) {
	proc := jsonReplyProc{}
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	w := &noFlushResponseWriter{}
	serveMCPPost(w, req, proc, HTTPConfig{
		PreferSSE: true,
		Logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
}

func TestParseInboundMessages_emptyBody(t *testing.T) {
	if _, err := ParseInboundMessages(nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunStdio_parseErrorThenOK(t *testing.T) {
	var in bytes.Buffer
	writeFramed(&in, []byte("{bad"))
	req := Message{JSONRPC: "2.0", ID: 1, Method: "ping"}
	b, _ := json.Marshal(req)
	writeFramed(&in, b)
	var out bytes.Buffer
	if err := RunStdio(context.Background(), mockProcessor{}, &in, &out); err != nil {
		t.Fatal(err)
	}
	if out.Len() == 0 {
		t.Fatal("expected response after parse error")
	}
}
