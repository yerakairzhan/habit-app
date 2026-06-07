// Package middleware contains Connect interceptors for auth, logging, and recovery.
package middleware

import (
	"context"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/rs/zerolog"
	"github.com/yourorg/habit-tracker/internal/auth"
	"github.com/yourorg/habit-tracker/internal/domain"
)

type contextKey string

const userIDKey contextKey = "user_id"

// publicProcedures is the set of Connect procedure names that skip auth.
// Format: /package.ServiceName/MethodName
var publicProcedures = map[string]bool{
	"/habit.v1.AuthService/Register":     true,
	"/habit.v1.AuthService/Login":        true,
	"/habit.v1.AuthService/RefreshToken": true,
}

// NewConnectAuthInterceptor returns a Connect interceptor that validates Bearer tokens.
func NewConnectAuthInterceptor(tokenSvc *auth.TokenService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if publicProcedures[req.Spec().Procedure] {
				return next(ctx, req)
			}

			// Try standard canonical header first
			authHeader := req.Header().Get("Authorization")
			if authHeader == "" {
				// Fallback to lowercase header forced by Envoy/HTTP2 proxies
				authHeader = req.Header().Get("authorization")
			}

			if authHeader == "" {
				return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrUnauthorized)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrTokenInvalid)
			}

			claims, err := tokenSvc.ValidateAccessToken(parts[1])
			if err != nil {
				if err == domain.ErrTokenExpired {
					return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrTokenExpired)
				}
				return nil, connect.NewError(connect.CodeUnauthenticated, domain.ErrTokenInvalid)
			}

			ctx = context.WithValue(ctx, userIDKey, claims.UserID)
			return next(ctx, req)
		}
	}
}

// NewConnectLoggingInterceptor returns a Connect interceptor that logs every call.
func NewConnectLoggingInterceptor(log zerolog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			resp, err := next(ctx, req)
			dur := time.Since(start)

			ev := log.Info()
			if err != nil {
				ev = log.Error().Err(err)
			}
			ev.
				Str("procedure", req.Spec().Procedure).
				Dur("duration", dur).
				Msg("rpc")

			return resp, err
		}
	}
}

// UserIDFromContext retrieves the authenticated user ID injected by the auth interceptor.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok
}
