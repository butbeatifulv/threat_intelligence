package mcpserver

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/butbeautifulv/veil/pkg/auth"
	"github.com/butbeautifulv/veil/pkg/mcp"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/usecase"
	playbookuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/playbook"
	procedureuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/procedure"
	frameworkuc "github.com/butbeautifulv/veil/knowledge/serve/internal/usecase/framework"
	"github.com/butbeautifulv/veil/knowledge/serve/internal/version"
)

type Server struct {
	uc        *usecase.ReadUsecase
	playbook  *playbookuc.Service
	procedure *procedureuc.Service
	framework *frameworkuc.Service
	auth      *auth.Stack
	logger    *slog.Logger
}

func NewServer(uc *usecase.ReadUsecase, playbook *playbookuc.Service, proc *procedureuc.Service, fw *frameworkuc.Service, stack *auth.Stack, logger *slog.Logger) *Server {
	if playbook == nil {
		playbook, _ = playbookuc.NewService()
	}
	if proc == nil {
		proc, _ = procedureuc.NewService()
	}
	if fw == nil {
		fw = frameworkuc.NewService()
	}
	return &Server{uc: uc, playbook: playbook, procedure: proc, framework: fw, auth: stack, logger: logger}
}

func (s *Server) Run(ctx context.Context, inReader any, outWriter any) error {
	return mcp.RunStdio(ctx, s, inReader, outWriter)
}

// ProcessMessage handles one JSON-RPC message. httpTransport selects protocol version and transport quirks.
func (s *Server) ProcessMessage(ctx context.Context, msg rpcMessage, httpTransport bool) (resp *rpcMessage, isNotification bool, err error) {
	result, rerr := s.handle(ctx, msg.Method, msg.Params, httpTransport)
	return mcp.BuildResponse(msg, result, rerr)
}

func (s *Server) handle(ctx context.Context, method string, params json.RawMessage, httpTransport bool) (any, error) {
	switch method {
	case "initialize":
		return mcp.InitializeResult(version.ServerName, version.MCP(), httpTransport, params), nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		return listToolsPayload(), nil

	case "tools/call":
		p, err := mcp.ParseToolCallParams(params)
		if err != nil {
			return nil, err
		}
		ctx, err = mcp.AuthorizeToolCall(ctx, s.auth, auth.PermGraphRead, auth.AuthorizeMCP)
		if err != nil {
			return nil, err
		}
		return s.callTool(ctx, p.Name, p.Arguments)

	case "notifications/initialized":
		return nil, nil
	}

	return nil, rpcErrf(codeMethodNotFound, "unknown method: %s", method)
}
