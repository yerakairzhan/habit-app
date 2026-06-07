// Package domain contains pure business entities and rules.
// It has NO external dependencies — no gRPC, no GORM, no database drivers.
package domain

import (
	"time"
)

// ─── User ────────────────────────────────────────────────────────────────────

type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ─── Habit ───────────────────────────────────────────────────────────────────

type ScheduleType string

const (
	ScheduleDaily        ScheduleType = "daily"
	ScheduleEveryOtherDay ScheduleType = "every_other_day"
	ScheduleWeekdays     ScheduleType = "weekdays"
)

// Weekday mirrors time.Weekday (0=Sunday … 6=Saturday).
type Weekday int

const (
	Sunday    Weekday = 0
	Monday    Weekday = 1
	Tuesday   Weekday = 2
	Wednesday Weekday = 3
	Thursday  Weekday = 4
	Friday    Weekday = 5
	Saturday  Weekday = 6
)

type Habit struct {
	ID           string
	UserID       string
	Title        string
	Color        string
	TargetPerDay int
	ScheduleType ScheduleType
	Weekdays     []Weekday
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// IsScheduledOn returns true if this habit should be tracked on the given date.
func (h *Habit) IsScheduledOn(date time.Time) bool {
	switch h.ScheduleType {
	case ScheduleDaily:
		return true

	case ScheduleEveryOtherDay:
		// Day 0 (creation date) is active; day 1 is not; day 2 is; etc.
		daysSince := int(date.Truncate(24*time.Hour).Sub(h.CreatedAt.Truncate(24*time.Hour)).Hours() / 24)
		if daysSince < 0 {
			return false
		}
		return daysSince%2 == 0

	case ScheduleWeekdays:
		today := Weekday(date.Weekday()) // time.Weekday and our Weekday align
		for _, wd := range h.Weekdays {
			if wd == today {
				return true
			}
		}
		return false
	}
	return false
}

// ─── HabitLog ────────────────────────────────────────────────────────────────

// HabitLog records daily progress for a single habit.
type HabitLog struct {
	ID        string
	HabitID   string
	UserID    string
	Date      time.Time // truncated to midnight UTC
	Progress  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Completed returns true when progress meets or exceeds target.
func (l *HabitLog) Completed(target int) bool {
	return l.Progress >= target
}

// ─── RefreshToken ────────────────────────────────────────────────────────────

type RefreshToken struct {
	ID        string
	UserID    string
	Token     string    // opaque UUID stored in DB
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (rt *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(rt.ExpiresAt)
}

// ─── Settings ────────────────────────────────────────────────────────────────

type Settings struct {
	ID                   string
	UserID               string
	Language             string
	NotificationsEnabled bool
	ShowDates            bool
	UpdatedAt            time.Time
}

// ─── Statistics ──────────────────────────────────────────────────────────────

type Stats struct {
	CurrentStreak      int
	BestStreak         int
	TrackedHabitsCount int
	CompletionRate30d  float32
}
