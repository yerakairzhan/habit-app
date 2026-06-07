package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/habit-tracker/internal/auth"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/service"
	"github.com/yourorg/habit-tracker/internal/testutil"
	"golang.org/x/crypto/bcrypt"
)

func newAuthService(
	users *testutil.MockUserRepository,
	tokens *testutil.MockRefreshTokenRepository,
	settings *testutil.MockSettingsRepository,
) *service.AuthService {
	tokenSvc := auth.NewTokenService("test-secret-key-that-is-32-chars!!")
	return service.NewAuthService(users, tokens, settings, tokenSvc)
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestAuthService_Register_Success(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	users.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "alice@example.com" && u.Name == "Alice"
	})).Run(func(args mock.Arguments) {
		u := args.Get(1).(*domain.User)
		u.ID = "user-123"
		u.CreatedAt = time.Now()
		u.UpdatedAt = time.Now()
	}).Return(nil)

	settings.On("GetOrCreate", mock.Anything, "user-123").
		Return(&domain.Settings{UserID: "user-123", Language: "en"}, nil)

	tokens.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).
		Run(func(args mock.Arguments) {
			rt := args.Get(1).(*domain.RefreshToken)
			rt.ID = "rt-1"
		}).Return(nil)

	user, pair, err := svc.Register(context.Background(), "alice@example.com", "password123", "Alice")

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	users.AssertExpectations(t)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	users.On("Create", mock.Anything, mock.Anything).Return(domain.ErrAlreadyExists)

	_, _, err := svc.Register(context.Background(), "dup@example.com", "password123", "Dup")
	assert.ErrorIs(t, err, domain.ErrAlreadyExists)
}

func TestAuthService_Register_ShortPassword(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	_, _, err := svc.Register(context.Background(), "a@b.com", "short", "Name")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	users.AssertNotCalled(t, "Create")
}

// ─── Login ───────────────────────────────────────────────────────────────────

func TestAuthService_Login_Success(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	existingUser := &domain.User{
		ID:           "user-1",
		Email:        "bob@example.com",
		PasswordHash: string(hash),
	}

	users.On("GetByEmail", mock.Anything, "bob@example.com").Return(existingUser, nil)
	tokens.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).
		Run(func(args mock.Arguments) {
			rt := args.Get(1).(*domain.RefreshToken)
			rt.ID = "rt-2"
		}).Return(nil)

	user, pair, err := svc.Login(context.Background(), "bob@example.com", "correctpass")

	require.NoError(t, err)
	assert.Equal(t, "user-1", user.ID)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correctpass"), bcrypt.MinCost)
	users.On("GetByEmail", mock.Anything, "bob@example.com").Return(&domain.User{
		ID:           "user-1",
		Email:        "bob@example.com",
		PasswordHash: string(hash),
	}, nil)

	_, _, err := svc.Login(context.Background(), "bob@example.com", "wrongpass")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	users.On("GetByEmail", mock.Anything, "ghost@example.com").Return(nil, domain.ErrNotFound)

	_, _, err := svc.Login(context.Background(), "ghost@example.com", "anypass")
	assert.ErrorIs(t, err, domain.ErrUnauthorized)
}

// ─── RefreshToken ────────────────────────────────────────────────────────────

func TestAuthService_RefreshToken_Success(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	existingRT := &domain.RefreshToken{
		ID:        "rt-old",
		UserID:    "user-1",
		Token:     "old-refresh-token",
		ExpiresAt: time.Now().Add(time.Hour),
	}
	tokens.On("GetByToken", mock.Anything, "old-refresh-token").Return(existingRT, nil)
	tokens.On("Delete", mock.Anything, "old-refresh-token").Return(nil)
	users.On("GetByID", mock.Anything, "user-1").Return(&domain.User{ID: "user-1", Email: "u@e.com"}, nil)
	tokens.On("Create", mock.Anything, mock.AnythingOfType("*domain.RefreshToken")).
		Run(func(args mock.Arguments) {
			rt := args.Get(1).(*domain.RefreshToken)
			rt.ID = "rt-new"
		}).Return(nil)

	pair, err := svc.RefreshToken(context.Background(), "old-refresh-token")
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
}

func TestAuthService_RefreshToken_Expired(t *testing.T) {
	users := &testutil.MockUserRepository{}
	tokens := &testutil.MockRefreshTokenRepository{}
	settings := &testutil.MockSettingsRepository{}
	svc := newAuthService(users, tokens, settings)

	expiredRT := &domain.RefreshToken{
		ID:        "rt-expired",
		UserID:    "user-1",
		Token:     "expired-token",
		ExpiresAt: time.Now().Add(-time.Hour), // already expired
	}
	tokens.On("GetByToken", mock.Anything, "expired-token").Return(expiredRT, nil)
	tokens.On("Delete", mock.Anything, "expired-token").Return(nil)

	_, err := svc.RefreshToken(context.Background(), "expired-token")
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
}
