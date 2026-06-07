// Package postgres contains GORM models and PostgreSQL repository implementations.
package postgres

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yourorg/habit-tracker/internal/domain"
	"gorm.io/gorm"
)

// ─── WeekdaySlice – custom GORM type ─────────────────────────────────────────

// WeekdaySlice stores []domain.Weekday as a JSON array in PostgreSQL (integer[]).
// We use JSONB for simplicity; an integer array column would also work.
type WeekdaySlice []domain.Weekday

func (w WeekdaySlice) Value() (driver.Value, error) {
	if w == nil {
		return "[]", nil
	}
	b, err := json.Marshal(w)
	return string(b), err
}

func (w *WeekdaySlice) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("cannot scan type %T into WeekdaySlice", value)
	}
	return json.Unmarshal(bytes, w)
}

// ─── UserModel ───────────────────────────────────────────────────────────────

type UserModel struct {
	ID           string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string `gorm:"uniqueIndex;not null"`
	Name         string `gorm:"not null"`
	PasswordHash string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

func (m *UserModel) ToDomain() *domain.User {
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		Name:         m.Name,
		PasswordHash: m.PasswordHash,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func UserModelFromDomain(u *domain.User) *UserModel {
	return &UserModel{
		ID:           u.ID,
		Email:        u.Email,
		Name:         u.Name,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// ─── HabitModel ──────────────────────────────────────────────────────────────

type HabitModel struct {
	ID           string       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string       `gorm:"type:uuid;not null;index"`
	Title        string       `gorm:"not null"`
	Color        string       `gorm:"not null;default:'#4F46E5'"`
	TargetPerDay int          `gorm:"not null;default:1"`
	ScheduleType string       `gorm:"type:varchar(32);not null"`
	Weekdays     WeekdaySlice `gorm:"type:jsonb;not null;default:'[]'"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"` // soft-delete
}

func (HabitModel) TableName() string { return "habits" }

func (m *HabitModel) ToDomain() *domain.Habit {
	return &domain.Habit{
		ID:           m.ID,
		UserID:       m.UserID,
		Title:        m.Title,
		Color:        m.Color,
		TargetPerDay: m.TargetPerDay,
		ScheduleType: domain.ScheduleType(m.ScheduleType),
		Weekdays:     []domain.Weekday(m.Weekdays),
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

func HabitModelFromDomain(h *domain.Habit) *HabitModel {
	return &HabitModel{
		ID:           h.ID,
		UserID:       h.UserID,
		Title:        h.Title,
		Color:        h.Color,
		TargetPerDay: h.TargetPerDay,
		ScheduleType: string(h.ScheduleType),
		Weekdays:     WeekdaySlice(h.Weekdays),
	}
}

// ─── HabitLogModel ───────────────────────────────────────────────────────────

type HabitLogModel struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	HabitID   string    `gorm:"type:uuid;not null;index:idx_habit_date,unique"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Date      time.Time `gorm:"type:date;not null;index:idx_habit_date,unique"`
	Progress  int       `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (HabitLogModel) TableName() string { return "habit_logs" }

func (m *HabitLogModel) ToDomain() *domain.HabitLog {
	return &domain.HabitLog{
		ID:        m.ID,
		HabitID:   m.HabitID,
		UserID:    m.UserID,
		Date:      m.Date,
		Progress:  m.Progress,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

func HabitLogModelFromDomain(l *domain.HabitLog) *HabitLogModel {
	return &HabitLogModel{
		ID:       l.ID,
		HabitID:  l.HabitID,
		UserID:   l.UserID,
		Date:     l.Date,
		Progress: l.Progress,
	}
}

// ─── RefreshTokenModel ───────────────────────────────────────────────────────

type RefreshTokenModel struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }

func (m *RefreshTokenModel) ToDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		CreatedAt: m.CreatedAt,
	}
}

func RefreshTokenModelFromDomain(rt *domain.RefreshToken) *RefreshTokenModel {
	return &RefreshTokenModel{
		ID:        rt.ID,
		UserID:    rt.UserID,
		Token:     rt.Token,
		ExpiresAt: rt.ExpiresAt,
	}
}

// ─── SettingsModel ───────────────────────────────────────────────────────────

type SettingsModel struct {
	ID                   string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID               string    `gorm:"type:uuid;uniqueIndex;not null"`
	Language             string    `gorm:"not null;default:'en'"`
	NotificationsEnabled bool      `gorm:"not null;default:true"`
	ShowDates            bool      `gorm:"not null;default:true"`
	UpdatedAt            time.Time
}

func (SettingsModel) TableName() string { return "user_settings" }

func (m *SettingsModel) ToDomain() *domain.Settings {
	return &domain.Settings{
		ID:                   m.ID,
		UserID:               m.UserID,
		Language:             m.Language,
		NotificationsEnabled: m.NotificationsEnabled,
		ShowDates:            m.ShowDates,
		UpdatedAt:            m.UpdatedAt,
	}
}

func SettingsModelFromDomain(s *domain.Settings) *SettingsModel {
	return &SettingsModel{
		ID:                   s.ID,
		UserID:               s.UserID,
		Language:             s.Language,
		NotificationsEnabled: s.NotificationsEnabled,
		ShowDates:            s.ShowDates,
	}
}
