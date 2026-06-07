// Package repository defines the persistence contracts used by the service layer.
// Concrete implementations live in repository/postgres.
package repository

import (
	"context"
	"time"

	"github.com/yourorg/habit-tracker/internal/domain"
)

// ─── User Repository ─────────────────────────────────────────────────────────

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}

// ─── Habit Repository ────────────────────────────────────────────────────────

type HabitFilter struct {
	UserID string
	Page   int // 1-indexed
	Size   int // items per page
}

type HabitRepository interface {
	Create(ctx context.Context, habit *domain.Habit) error
	GetByID(ctx context.Context, id string) (*domain.Habit, error)
	// GetByIDForUser returns ErrForbidden when habit exists but belongs to another user.
	GetByIDForUser(ctx context.Context, id, userID string) (*domain.Habit, error)
	Update(ctx context.Context, habit *domain.Habit) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter HabitFilter) ([]*domain.Habit, int64, error)
	// ListForUser returns all habits owned by a user (no pagination) — used for
	// today-view and stats where we need the full set.
	ListForUser(ctx context.Context, userID string) ([]*domain.Habit, error)
}

// ─── HabitLog Repository ─────────────────────────────────────────────────────

type HabitLogRepository interface {
	// Upsert creates or returns the existing log row for (habitID, date).
	Upsert(ctx context.Context, log *domain.HabitLog) error
	GetByHabitAndDate(ctx context.Context, habitID string, date time.Time) (*domain.HabitLog, error)
	// IncrementProgress atomically increments and returns updated log.
	IncrementProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error)
	// DecrementProgress atomically decrements (floor 0) and returns updated log.
	DecrementProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error)
	// ResetProgress sets progress to 0 and returns updated log.
	ResetProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error)
	// ListByHabitSince returns logs for a habit on or after `since` (for streak calc).
	ListByHabitSince(ctx context.Context, habitID string, since time.Time) ([]*domain.HabitLog, error)
	// ListByUserSince returns logs for all habits of a user since the given date.
	ListByUserSince(ctx context.Context, userID string, since time.Time) ([]*domain.HabitLog, error)
	// CountTrackedHabits returns how many habits have at least one completed log entry.
	CountTrackedHabits(ctx context.Context, userID string) (int, error)
}

// ─── RefreshToken Repository ─────────────────────────────────────────────────

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Delete(ctx context.Context, token string) error
	DeleteAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) error
}

// ─── Settings Repository ─────────────────────────────────────────────────────

type SettingsRepository interface {
	// GetOrCreate returns existing settings or creates defaults for the user.
	GetOrCreate(ctx context.Context, userID string) (*domain.Settings, error)
	Update(ctx context.Context, s *domain.Settings) error
}
