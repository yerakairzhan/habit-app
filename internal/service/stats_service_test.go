package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/service"
	"github.com/yourorg/habit-tracker/internal/testutil"
)

func dateAt(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func makeLog(habitID string, date time.Time, progress int) *domain.HabitLog {
	return &domain.HabitLog{
		ID:       "log-" + date.Format("20060102"),
		HabitID:  habitID,
		UserID:   "user-1",
		Date:     date,
		Progress: progress,
	}
}

func newStatsService(habits *testutil.MockHabitRepository, logs *testutil.MockHabitLogRepository) *service.StatsService {
	return service.NewStatsService(habits, logs)
}

// TestStatsService_Streak_AllCompleted verifies a perfect streak across 5 days.
func TestStatsService_Streak_AllCompleted(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newStatsService(habits, logs)

	createdAt := dateAt(2024, 1, 1)
	h := &domain.Habit{
		ID:           "h1",
		UserID:       "user-1",
		ScheduleType: domain.ScheduleDaily,
		TargetPerDay: 1,
		CreatedAt:    createdAt,
	}

	// 5 days of completed logs
	logList := []*domain.HabitLog{
		makeLog("h1", dateAt(2024, 1, 1), 1),
		makeLog("h1", dateAt(2024, 1, 2), 1),
		makeLog("h1", dateAt(2024, 1, 3), 1),
		makeLog("h1", dateAt(2024, 1, 4), 1),
		makeLog("h1", dateAt(2024, 1, 5), 1),
	}

	habits.On("GetByIDForUser", mock.Anything, "h1", "user-1").Return(h, nil)
	logs.On("ListByHabitSince", mock.Anything, "h1", mock.Anything).Return(logList, nil)
	logs.On("CountTrackedHabits", mock.Anything, "user-1").Return(1, nil)

	// We test the internal streak logic indirectly via GetStats.
	// For a clean test we need today to match the last log date.
	// Since the streak algo uses time.Now() internally we can only assert
	// that best_streak >= 5 when all days were completed.
	stats, err := svc.GetStats(context.Background(), "user-1", "h1")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.BestStreak, 5)
	assert.Equal(t, 1, stats.TrackedHabitsCount)
}

// TestStatsService_Streak_BrokenByMissedDay verifies streak resets after a missed day.
func TestStatsService_Streak_BrokenByMissedDay(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newStatsService(habits, logs)

	createdAt := dateAt(2024, 1, 1)
	h := &domain.Habit{
		ID:           "h1",
		UserID:       "user-1",
		ScheduleType: domain.ScheduleDaily,
		TargetPerDay: 1,
		CreatedAt:    createdAt,
	}

	// Days 1,2 done. Day 3 missed. Days 4,5 done.
	logList := []*domain.HabitLog{
		makeLog("h1", dateAt(2024, 1, 1), 1),
		makeLog("h1", dateAt(2024, 1, 2), 1),
		// 2024-01-03 missing → streak break
		makeLog("h1", dateAt(2024, 1, 4), 1),
		makeLog("h1", dateAt(2024, 1, 5), 1),
	}

	habits.On("GetByIDForUser", mock.Anything, "h1", "user-1").Return(h, nil)
	logs.On("ListByHabitSince", mock.Anything, "h1", mock.Anything).Return(logList, nil)
	logs.On("CountTrackedHabits", mock.Anything, "user-1").Return(1, nil)

	stats, err := svc.GetStats(context.Background(), "user-1", "h1")
	require.NoError(t, err)
	// best streak is 2 (days 1-2 OR days 4-5); current streak is broken.
	assert.Equal(t, 2, stats.BestStreak)
	assert.Equal(t, 0, stats.CurrentStreak)
}

// TestStatsService_EveryOtherDay_NonScheduledDaysIgnored verifies that
// non-scheduled days do not break the streak.
func TestStatsService_EveryOtherDay_NonScheduledDaysIgnored(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newStatsService(habits, logs)

	createdAt := dateAt(2024, 1, 1)
	h := &domain.Habit{
		ID:           "h1",
		UserID:       "user-1",
		ScheduleType: domain.ScheduleEveryOtherDay,
		TargetPerDay: 1,
		CreatedAt:    createdAt,
	}

	// Scheduled: Jan 1, 3, 5. All completed.
	logList := []*domain.HabitLog{
		makeLog("h1", dateAt(2024, 1, 1), 1),
		makeLog("h1", dateAt(2024, 1, 3), 1),
		makeLog("h1", dateAt(2024, 1, 5), 1),
	}

	habits.On("GetByIDForUser", mock.Anything, "h1", "user-1").Return(h, nil)
	logs.On("ListByHabitSince", mock.Anything, "h1", mock.Anything).Return(logList, nil)
	logs.On("CountTrackedHabits", mock.Anything, "user-1").Return(1, nil)

	stats, err := svc.GetStats(context.Background(), "user-1", "h1")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.BestStreak, 3)
}

// TestStatsService_NoHabits returns zeroed stats cleanly.
func TestStatsService_NoHabits(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newStatsService(habits, logs)

	habits.On("ListForUser", mock.Anything, "user-1").Return([]*domain.Habit{}, nil)
	logs.On("ListByUserSince", mock.Anything, "user-1", mock.Anything).Return([]*domain.HabitLog{}, nil)
	logs.On("CountTrackedHabits", mock.Anything, "user-1").Return(0, nil)

	stats, err := svc.GetStats(context.Background(), "user-1", "")
	require.NoError(t, err)
	assert.Equal(t, 0, stats.CurrentStreak)
	assert.Equal(t, 0, stats.BestStreak)
	assert.Equal(t, 0, stats.TrackedHabitsCount)
	assert.Equal(t, float32(0), stats.CompletionRate30d)
}
