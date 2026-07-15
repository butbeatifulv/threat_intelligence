package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

// FramedRW reads/writes LSP-style Content-Length framed JSON-RPC messages.
type FramedRW struct {
	r *bufio.Reader
	w io.Writer
	m sync.Mutex
}

// NewFramedRW constructs a framed reader/writer over in/out.
func NewFramedRW(r io.Reader, w io.Writer) *FramedRW {
	return &FramedRW{r: bufio.NewReader(r), w: w}
}

func (rw *FramedRW) Read(ctx context.Context) ([]byte, error) {
	type result struct {
		buf []byte
		err error
	}
	ch := make(chan result, 1)
	go func() {
		buf, err := rw.readFrame()
		ch <- result{buf: buf, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.buf, r.err
	}
}

func (rw *FramedRW) readFrame() ([]byte, error) {
	var contentLen int
	for {
		line, err := rw.r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(strings.ToLower(parts[0]))
		v := strings.TrimSpace(parts[1])
		if k == "content-length" {
			n, _ := strconv.Atoi(v)
			contentLen = n
		}
	}
	if contentLen <= 0 {
		return nil, fmt.Errorf("missing/invalid Content-Length")
	}
	buf := make([]byte, contentLen)
	if _, err := io.ReadFull(rw.r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (rw *FramedRW) WriteJSON(ctx context.Context, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.m.Lock()
	defer rw.m.Unlock()
	var out bytes.Buffer
	_, _ = fmt.Fprintf(&out, "Content-Length: %d\r\n\r\n", len(b))
	out.Write(b)
	_, err = rw.w.Write(out.Bytes())
	return err
}
