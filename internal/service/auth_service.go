// Package service contains all application use-cases.
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/habit-tracker/internal/auth"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles registration, login, and token lifecycle.
type AuthService struct {
	users         repository.UserRepository
	refreshTokens repository.RefreshTokenRepository
	settings      repository.SettingsRepository
	tokenSvc      *auth.TokenService
}

func NewAuthService(
	users repository.UserRepository,
	refreshTokens repository.RefreshTokenRepository,
	settings repository.SettingsRepository,
	tokenSvc *auth.TokenService,
) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: refreshTokens,
		settings:      settings,
		tokenSvc:      tokenSvc,
	}
}

// TokenPair groups access + refresh tokens for return values.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// Register creates a new user, seeds default settings, and returns a token pair.
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*domain.User, TokenPair, error) {
	if email == "" || password == "" || name == "" {
		return nil, TokenPair{}, domain.ErrInvalidInput
	}
	if len(password) < 8 {
		return nil, TokenPair{}, domain.ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, TokenPair{}, domain.ErrInternal
	}

	user := &domain.User{
		Email:        email,
		Name:         name,
		PasswordHash: string(hash),
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, TokenPair{}, err // ErrAlreadyExists propagates
	}

	// Seed default settings.
	if _, err := s.settings.GetOrCreate(ctx, user.ID); err != nil {
		// Non-fatal: log in production, continue.
		_ = err
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

// Login validates credentials and returns a fresh token pair.
func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, TokenPair, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, TokenPair{}, domain.ErrUnauthorized
		}
		return nil, TokenPair{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, TokenPair{}, domain.ErrUnauthorized
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, TokenPair{}, err
	}
	return user, pair, nil
}

// RefreshToken rotates the refresh token and issues a new access token.
// The old refresh token is deleted (one-time use).
func (s *AuthService) RefreshToken(ctx context.Context, rawRefresh string) (TokenPair, error) {
	rt, err := s.refreshTokens.GetByToken(ctx, rawRefresh)
	if err != nil {
		return TokenPair{}, domain.ErrTokenInvalid
	}
	if rt.IsExpired() {
		_ = s.refreshTokens.Delete(ctx, rawRefresh) // clean up eagerly
		return TokenPair{}, domain.ErrTokenExpired
	}

	// Rotate: delete old token before issuing new one.
	if err := s.refreshTokens.Delete(ctx, rawRefresh); err != nil {
		return TokenPair{}, domain.ErrInternal
	}

	user, err := s.users.GetByID(ctx, rt.UserID)
	if err != nil {
		return TokenPair{}, err
	}

	return s.issueTokenPair(ctx, user)
}

// Logout revokes the refresh token.
func (s *AuthService) Logout(ctx context.Context, rawRefresh string) error {
	return s.refreshTokens.Delete(ctx, rawRefresh)
}

// ─── Private helpers ─────────────────────────────────────────────────────────

func (s *AuthService) issueTokenPair(ctx context.Context, user *domain.User) (TokenPair, error) {
	accessToken, err := s.tokenSvc.IssueAccessToken(user)
	if err != nil {
		return TokenPair{}, domain.ErrInternal
	}

	rt := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     uuid.NewString(),
		ExpiresAt: time.Now().UTC().Add(auth.RefreshTokenDuration),
	}
	if err := s.refreshTokens.Create(ctx, rt); err != nil {
		return TokenPair{}, domain.ErrInternal
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rt.Token,
	}, nil
}
