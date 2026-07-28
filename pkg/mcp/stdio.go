package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// RunStdio serves MCP over Content-Length framed stdio until ctx is done or EOF.
func RunStdio(ctx context.Context, proc Processor, inReader, outWriter any) error {
	in, ok := inReader.(interface{ Read([]byte) (int, error) })
	if !ok {
		return fmt.Errorf("invalid stdin reader")
	}
	out, ok := outWriter.(interface{ Write([]byte) (int, error) })
	if !ok {
		return fmt.Errorf("invalid stdout writer")
	}

	rw := NewFramedRW(in, out)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		payload, err := rw.Read(ctx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		var msg Message
		if err := json.Unmarshal(payload, &msg); err != nil {
			_ = rw.WriteJSON(ctx, Message{
				JSONRPC: "2.0",
				ID:      nil,
				Error:   &RPCError{Code: CodeParseError, Message: "Parse error"},
			})
			continue
		}

		if msg.Method == "" {
			continue
		}

		resp, isNotification, perr := proc.ProcessMessage(ctx, msg, false)
		if perr != nil {
			return perr
		}
		if isNotification || resp == nil {
			continue
		}
		if err := rw.WriteJSON(ctx, *resp); err != nil {
			return err
		}
	}
}
