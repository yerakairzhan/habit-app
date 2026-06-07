package domain

import "errors"

// Sentinel errors that service and handler layers translate to gRPC status codes.
var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
	ErrTokenExpired      = errors.New("token expired")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrPasswordMismatch  = errors.New("password mismatch")
	ErrInternal          = errors.New("internal error")
)
