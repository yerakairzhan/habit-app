package handler

import (
	"context"

	"connectrpc.com/connect"
	pb "github.com/yourorg/habit-tracker/gen/habit/v1"
	"github.com/yourorg/habit-tracker/gen/habit/v1/habitv1connect"
	"github.com/yourorg/habit-tracker/internal/service"
)

// AuthHandler implements habitv1connect.AuthServiceHandler.
type AuthHandler struct {
	svc *service.AuthService
}

// Compile-time check.
var _ habitv1connect.AuthServiceHandler = (*AuthHandler)(nil)

func NewAuthHandler(svc *service.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(
	ctx context.Context,
	req *connect.Request[pb.RegisterRequest],
) (*connect.Response[pb.RegisterResponse], error) {
	user, pair, err := h.svc.Register(ctx, req.Msg.Email, req.Msg.Password, req.Msg.Name)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.RegisterResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         &pb.User{Id: user.ID, Email: user.Email, Name: user.Name},
	}), nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *connect.Request[pb.LoginRequest],
) (*connect.Response[pb.LoginResponse], error) {
	user, pair, err := h.svc.Login(ctx, req.Msg.Email, req.Msg.Password)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		User:         &pb.User{Id: user.ID, Email: user.Email, Name: user.Name},
	}), nil
}

func (h *AuthHandler) RefreshToken(
	ctx context.Context,
	req *connect.Request[pb.RefreshTokenRequest],
) (*connect.Response[pb.RefreshTokenResponse], error) {
	pair, err := h.svc.RefreshToken(ctx, req.Msg.RefreshToken)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
	}), nil
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	req *connect.Request[pb.LogoutRequest],
) (*connect.Response[pb.LogoutResponse], error) {
	if err := h.svc.Logout(ctx, req.Msg.RefreshToken); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.LogoutResponse{}), nil
}
