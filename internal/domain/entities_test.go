package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yourorg/habit-tracker/internal/domain"
)

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func TestHabit_IsScheduledOn_Daily(t *testing.T) {
	h := &domain.Habit{
		ScheduleType: domain.ScheduleDaily,
		CreatedAt:    date(2024, 1, 1),
	}
	assert.True(t, h.IsScheduledOn(date(2024, 1, 1)))
	assert.True(t, h.IsScheduledOn(date(2024, 6, 15)))
	assert.True(t, h.IsScheduledOn(date(2025, 12, 31)))
}

func TestHabit_IsScheduledOn_EveryOtherDay(t *testing.T) {
	h := &domain.Habit{
		ScheduleType: domain.ScheduleEveryOtherDay,
		CreatedAt:    date(2024, 1, 1),
	}
	assert.True(t, h.IsScheduledOn(date(2024, 1, 1)))  // day 0 → active
	assert.False(t, h.IsScheduledOn(date(2024, 1, 2))) // day 1 → skip
	assert.True(t, h.IsScheduledOn(date(2024, 1, 3)))  // day 2 → active
	assert.False(t, h.IsScheduledOn(date(2024, 1, 4))) // day 3 → skip
	// before creation date
	assert.False(t, h.IsScheduledOn(date(2023, 12, 31)))
}

func TestHabit_IsScheduledOn_Weekdays(t *testing.T) {
	h := &domain.Habit{
		ScheduleType: domain.ScheduleWeekdays,
		Weekdays:     []domain.Weekday{domain.Monday, domain.Wednesday, domain.Friday},
		CreatedAt:    date(2024, 1, 1),
	}
	// 2024-01-15 is Monday
	assert.True(t, h.IsScheduledOn(date(2024, 1, 15)))
	// 2024-01-16 is Tuesday
	assert.False(t, h.IsScheduledOn(date(2024, 1, 16)))
	// 2024-01-17 is Wednesday
	assert.True(t, h.IsScheduledOn(date(2024, 1, 17)))
	// 2024-01-20 is Saturday
	assert.False(t, h.IsScheduledOn(date(2024, 1, 20)))
}

func TestHabitLog_Completed(t *testing.T) {
	log := &domain.HabitLog{Progress: 3}
	assert.True(t, log.Completed(3))
	assert.True(t, log.Completed(2))
	assert.False(t, log.Completed(4))
}

func TestRefreshToken_IsExpired(t *testing.T) {
	expired := &domain.RefreshToken{ExpiresAt: time.Now().UTC().Add(-time.Minute)}
	assert.True(t, expired.IsExpired())

	valid := &domain.RefreshToken{ExpiresAt: time.Now().UTC().Add(time.Hour)}
	assert.False(t, valid.IsExpired())
}
