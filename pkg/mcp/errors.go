package mcp

import (
	"errors"
	"fmt"

	"github.com/butbeautifulv/veil/pkg/auth"
)

func Err(code int, msg string) error {
	return &RPCError{Code: code, Message: msg}
}

func Errf(code int, format string, args ...any) error {
	return Err(code, fmt.Sprintf(format, args...))
}

// ToRPCError maps errors to JSON-RPC errors (auth-aware).
func ToRPCError(err error) *RPCError {
	var re *RPCError
	if errors.As(err, &re) {
		return re
	}
	if errors.Is(err, auth.ErrForbidden) {
		return &RPCError{Code: CodeAuthError, Message: "forbidden"}
	}
	if errors.Is(err, auth.ErrUnauthorized) {
		return &RPCError{Code: CodeAuthError, Message: "unauthorized"}
	}
	return &RPCError{Code: CodeInternal, Message: err.Error()}
}
