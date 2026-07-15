package mcpserver

import (
	"io"
	"net/http"

	"github.com/butbeautifulv/veil/pkg/mcp"
)

type rpcMessage = mcp.Message

const (
	codeParseError      = mcp.CodeParseError
	codeInvalidRequest  = mcp.CodeInvalidRequest
	codeMethodNotFound  = mcp.CodeMethodNotFound
	codeInvalidParams   = mcp.CodeInvalidParams
	codeInternal        = mcp.CodeInternal
	codeToolError       = mcp.CodeToolError
	codeAuthError       = mcp.CodeAuthError
	protocolVersionHTTP = mcp.ProtocolVersionHTTP
	defaultProtocol     = mcp.DefaultProtocol
	protocol20241105    = mcp.Protocol20241105
	protocol20250326    = mcp.Protocol20250326
)

func newFramedRW(r io.Reader, w io.Writer) *mcp.FramedRW {
	return mcp.NewFramedRW(r, w)
}

func wantsSSE(r *http.Request, preferSSE bool) bool {
	return mcp.WantsSSE(r, preferSSE)
}

func rpcErrf(code int, format string, args ...any) error {
	return mcp.Errf(code, format, args...)
}

func negotiateProtocol(params []byte) string {
	return mcp.NegotiateProtocol(params)
}
