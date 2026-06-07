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

func newHabitService(habits *testutil.MockHabitRepository, logs *testutil.MockHabitLogRepository) *service.HabitService {
	return service.NewHabitService(habits, logs)
}

// ─── CreateHabit ─────────────────────────────────────────────────────────────

func TestHabitService_CreateHabit_Success(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	habits.On("Create", mock.Anything, mock.MatchedBy(func(h *domain.Habit) bool {
		return h.UserID == "user-1" && h.Title == "Morning Run"
	})).Run(func(args mock.Arguments) {
		h := args.Get(1).(*domain.Habit)
		h.ID = "habit-1"
		h.CreatedAt = time.Now()
		h.UpdatedAt = time.Now()
	}).Return(nil)

	habit := &domain.Habit{
		Title:        "Morning Run",
		Color:        "#FF5733",
		TargetPerDay: 1,
		ScheduleType: domain.ScheduleDaily,
	}

	created, err := svc.CreateHabit(context.Background(), "user-1", habit)
	require.NoError(t, err)
	assert.Equal(t, "habit-1", created.ID)
	assert.Equal(t, "user-1", created.UserID)
	habits.AssertExpectations(t)
}

func TestHabitService_CreateHabit_InvalidTitle(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	habit := &domain.Habit{
		Title:        "", // empty title
		TargetPerDay: 1,
		ScheduleType: domain.ScheduleDaily,
	}

	_, err := svc.CreateHabit(context.Background(), "user-1", habit)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
	habits.AssertNotCalled(t, "Create")
}

func TestHabitService_CreateHabit_WeekdaysWithoutDays(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	habit := &domain.Habit{
		Title:        "Yoga",
		TargetPerDay: 1,
		ScheduleType: domain.ScheduleWeekdays,
		Weekdays:     []domain.Weekday{}, // empty weekdays for weekdays schedule
	}

	_, err := svc.CreateHabit(context.Background(), "user-1", habit)
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

// ─── DeleteHabit ─────────────────────────────────────────────────────────────

func TestHabitService_DeleteHabit_Success(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-1").
		Return(&domain.Habit{ID: "habit-1", UserID: "user-1"}, nil)
	habits.On("Delete", mock.Anything, "habit-1").Return(nil)

	err := svc.DeleteHabit(context.Background(), "user-1", "habit-1")
	require.NoError(t, err)
	habits.AssertExpectations(t)
}

func TestHabitService_DeleteHabit_Forbidden(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-2").
		Return(nil, domain.ErrForbidden)

	err := svc.DeleteHabit(context.Background(), "user-2", "habit-1")
	assert.ErrorIs(t, err, domain.ErrForbidden)
	habits.AssertNotCalled(t, "Delete")
}

// ─── GetTodaysHabits ─────────────────────────────────────────────────────────

func TestHabitService_GetTodaysHabits_FiltersSchedule(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	today := time.Now().UTC()

	// Weekday habit that does NOT include today's weekday.
	// Build a list of weekdays that excludes today.
	todayWD := domain.Weekday(today.Weekday())
	var excludeToday []domain.Weekday
	for wd := domain.Sunday; wd <= domain.Saturday; wd++ {
		if wd != todayWD {
			excludeToday = append(excludeToday, wd)
			break // just one day that isn't today
		}
	}

	dailyHabit := &domain.Habit{
		ID: "h1", UserID: "user-1",
		Title:        "Daily",
		TargetPerDay: 1,
		ScheduleType: domain.ScheduleDaily,
		CreatedAt:    today.AddDate(0, 0, -5),
	}
	skippedHabit := &domain.Habit{
		ID: "h2", UserID: "user-1",
		Title:        "Not today",
		TargetPerDay: 1,
		ScheduleType: domain.ScheduleWeekdays,
		Weekdays:     excludeToday,
		CreatedAt:    today.AddDate(0, 0, -5),
	}

	habits.On("ListForUser", mock.Anything, "user-1").
		Return([]*domain.Habit{dailyHabit, skippedHabit}, nil)
	logs.On("GetByHabitAndDate", mock.Anything, "h1", mock.Anything).
		Return(nil, domain.ErrNotFound)

	views, err := svc.GetTodaysHabits(context.Background(), "user-1")
	require.NoError(t, err)
	// Only the daily habit should appear.
	require.Len(t, views, 1)
	assert.Equal(t, "h1", views[0].Habit.ID)
	assert.Equal(t, 0, views[0].Progress)
	assert.False(t, views[0].Completed)
}

func TestHabitService_GetTodaysHabits_WithProgress(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newHabitService(habits, logs)

	today := time.Now().UTC()
	h := &domain.Habit{
		ID: "h1", UserID: "user-1",
		Title: "Drink Water", TargetPerDay: 3,
		ScheduleType: domain.ScheduleDaily,
		CreatedAt:    today.AddDate(0, 0, -1),
	}

	habits.On("ListForUser", mock.Anything, "user-1").Return([]*domain.Habit{h}, nil)
	logs.On("GetByHabitAndDate", mock.Anything, "h1", mock.Anything).
		Return(&domain.HabitLog{HabitID: "h1", Progress: 3}, nil)

	views, err := svc.GetTodaysHabits(context.Background(), "user-1")
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Equal(t, 3, views[0].Progress)
	assert.Equal(t, 3, views[0].Target)
	assert.True(t, views[0].Completed)
}
