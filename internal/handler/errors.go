// Package handler contains Connect service implementations.
package handler

import (
	"connectrpc.com/connect"
	"github.com/yourorg/habit-tracker/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toConnectError maps domain sentinel errors to Connect status codes.
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	switch err {
	case domain.ErrNotFound:
		return connect.NewError(connect.CodeNotFound, err)
	case domain.ErrAlreadyExists:
		return connect.NewError(connect.CodeAlreadyExists, err)
	case domain.ErrInvalidInput:
		return connect.NewError(connect.CodeInvalidArgument, err)
	case domain.ErrUnauthorized:
		return connect.NewError(connect.CodeUnauthenticated, err)
	case domain.ErrForbidden:
		return connect.NewError(connect.CodePermissionDenied, err)
	case domain.ErrTokenExpired:
		return connect.NewError(connect.CodeUnauthenticated, domain.ErrTokenExpired)
	case domain.ErrTokenInvalid:
		return connect.NewError(connect.CodeUnauthenticated, domain.ErrTokenInvalid)
	case domain.ErrPasswordMismatch:
		return connect.NewError(connect.CodeUnauthenticated, domain.ErrPasswordMismatch)
	default:
		return connect.NewError(connect.CodeInternal, domain.ErrInternal)
	}
}

// toGRPCError kept for backward-compat; routes through connect errors.
func toGRPCError(err error) error { return toConnectError(err) }

// Keep gRPC status import satisfied for any lingering uses.
var _ = status.Error
var _ = codes.OK
