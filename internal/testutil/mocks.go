// Package testutil provides mock implementations of repository interfaces
// for unit-testing the service layer without a real database.
package testutil

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
)

// ─── MockUserRepository ──────────────────────────────────────────────────────

type MockUserRepository struct{ mock.Mock }

var _ repository.UserRepository = (*MockUserRepository)(nil)

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if u, ok := args.Get(0).(*domain.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if u, ok := args.Get(0).(*domain.User); ok {
		return u, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

// ─── MockHabitRepository ─────────────────────────────────────────────────────

type MockHabitRepository struct{ mock.Mock }

var _ repository.HabitRepository = (*MockHabitRepository)(nil)

func (m *MockHabitRepository) Create(ctx context.Context, h *domain.Habit) error {
	return m.Called(ctx, h).Error(0)
}
func (m *MockHabitRepository) GetByID(ctx context.Context, id string) (*domain.Habit, error) {
	args := m.Called(ctx, id)
	if h, ok := args.Get(0).(*domain.Habit); ok {
		return h, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockHabitRepository) GetByIDForUser(ctx context.Context, id, userID string) (*domain.Habit, error) {
	args := m.Called(ctx, id, userID)
	if h, ok := args.Get(0).(*domain.Habit); ok {
		return h, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockHabitRepository) Update(ctx context.Context, h *domain.Habit) error {
	return m.Called(ctx, h).Error(0)
}
func (m *MockHabitRepository) Delete(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockHabitRepository) List(ctx context.Context, f repository.HabitFilter) ([]*domain.Habit, int64, error) {
	args := m.Called(ctx, f)
	habits, _ := args.Get(0).([]*domain.Habit)
	return habits, int64(args.Int(1)), args.Error(2)
}
func (m *MockHabitRepository) ListForUser(ctx context.Context, userID string) ([]*domain.Habit, error) {
	args := m.Called(ctx, userID)
	habits, _ := args.Get(0).([]*domain.Habit)
	return habits, args.Error(1)
}

// ─── MockHabitLogRepository ──────────────────────────────────────────────────

type MockHabitLogRepository struct{ mock.Mock }

var _ repository.HabitLogRepository = (*MockHabitLogRepository)(nil)

func (m *MockHabitLogRepository) Upsert(ctx context.Context, log *domain.HabitLog) error {
	return m.Called(ctx, log).Error(0)
}
func (m *MockHabitLogRepository) GetByHabitAndDate(ctx context.Context, habitID string, date time.Time) (*domain.HabitLog, error) {
	args := m.Called(ctx, habitID, date)
	if l, ok := args.Get(0).(*domain.HabitLog); ok {
		return l, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockHabitLogRepository) IncrementProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error) {
	args := m.Called(ctx, habitID, userID, date)
	if l, ok := args.Get(0).(*domain.HabitLog); ok {
		return l, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockHabitLogRepository) DecrementProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error) {
	args := m.Called(ctx, habitID, userID, date)
	if l, ok := args.Get(0).(*domain.HabitLog); ok {
		return l, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockHabitLogRepository) ResetProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error) {
	args := m.Called(ctx, habitID, userID, date)
	if l, ok := args.Get(0).(*domain.HabitLog); ok {
		return l, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockHabitLogRepository) ListByHabitSince(ctx context.Context, habitID string, since time.Time) ([]*domain.HabitLog, error) {
	args := m.Called(ctx, habitID, since)
	logs, _ := args.Get(0).([]*domain.HabitLog)
	return logs, args.Error(1)
}
func (m *MockHabitLogRepository) ListByUserSince(ctx context.Context, userID string, since time.Time) ([]*domain.HabitLog, error) {
	args := m.Called(ctx, userID, since)
	logs, _ := args.Get(0).([]*domain.HabitLog)
	return logs, args.Error(1)
}
func (m *MockHabitLogRepository) CountTrackedHabits(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// ─── MockRefreshTokenRepository ──────────────────────────────────────────────

type MockRefreshTokenRepository struct{ mock.Mock }

var _ repository.RefreshTokenRepository = (*MockRefreshTokenRepository)(nil)

func (m *MockRefreshTokenRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	return m.Called(ctx, rt).Error(0)
}
func (m *MockRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	if rt, ok := args.Get(0).(*domain.RefreshToken); ok {
		return rt, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRefreshTokenRepository) Delete(ctx context.Context, token string) error {
	return m.Called(ctx, token).Error(0)
}
func (m *MockRefreshTokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *MockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// ─── MockSettingsRepository ──────────────────────────────────────────────────

type MockSettingsRepository struct{ mock.Mock }

var _ repository.SettingsRepository = (*MockSettingsRepository)(nil)

func (m *MockSettingsRepository) GetOrCreate(ctx context.Context, userID string) (*domain.Settings, error) {
	args := m.Called(ctx, userID)
	if s, ok := args.Get(0).(*domain.Settings); ok {
		return s, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockSettingsRepository) Update(ctx context.Context, s *domain.Settings) error {
	return m.Called(ctx, s).Error(0)
}
