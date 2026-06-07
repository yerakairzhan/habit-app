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

func newProgressService(habits *testutil.MockHabitRepository, logs *testutil.MockHabitLogRepository) *service.ProgressService {
	return service.NewProgressService(habits, logs)
}

var testHabit = &domain.Habit{
	ID:           "habit-1",
	UserID:       "user-1",
	TargetPerDay: 3,
	ScheduleType: domain.ScheduleDaily,
}

func TestProgressService_Increment_Success(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newProgressService(habits, logs)

	today := time.Now().UTC()

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-1").Return(testHabit, nil)
	logs.On("IncrementProgress", mock.Anything, "habit-1", "user-1", mock.MatchedBy(func(d time.Time) bool {
		return d.UTC().Format("2006-01-02") == today.Format("2006-01-02")
	})).Return(&domain.HabitLog{HabitID: "habit-1", Progress: 1}, nil)

	res, err := svc.IncrementProgress(context.Background(), "user-1", "habit-1", "")
	require.NoError(t, err)
	assert.Equal(t, 1, res.Progress)
	assert.Equal(t, 3, res.Target)
	assert.False(t, res.Completed)
}

func TestProgressService_Increment_Completes(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newProgressService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-1").Return(testHabit, nil)
	logs.On("IncrementProgress", mock.Anything, "habit-1", "user-1", mock.Anything).
		Return(&domain.HabitLog{HabitID: "habit-1", Progress: 3}, nil)

	res, err := svc.IncrementProgress(context.Background(), "user-1", "habit-1", "")
	require.NoError(t, err)
	assert.True(t, res.Completed) // 3 >= 3
}

func TestProgressService_Decrement_FloorZero(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newProgressService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-1").Return(testHabit, nil)
	// DB returns 0 after GREATEST(0, progress - 1) on an already-zero row.
	logs.On("DecrementProgress", mock.Anything, "habit-1", "user-1", mock.Anything).
		Return(&domain.HabitLog{HabitID: "habit-1", Progress: 0}, nil)

	res, err := svc.DecrementProgress(context.Background(), "user-1", "habit-1", "")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Progress)
}

func TestProgressService_InvalidDate(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newProgressService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-1").Return(testHabit, nil)

	_, err := svc.IncrementProgress(context.Background(), "user-1", "habit-1", "not-a-date")
	assert.ErrorIs(t, err, domain.ErrInvalidInput)
}

func TestProgressService_GetProgress_NoLogYet(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newProgressService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-1").Return(testHabit, nil)
	logs.On("GetByHabitAndDate", mock.Anything, "habit-1", mock.Anything).
		Return(nil, domain.ErrNotFound)

	res, err := svc.GetProgress(context.Background(), "user-1", "habit-1", "")
	require.NoError(t, err)
	assert.Equal(t, 0, res.Progress)
	assert.Equal(t, 3, res.Target)
	assert.False(t, res.Completed)
}

func TestProgressService_ForbiddenHabit(t *testing.T) {
	habits := &testutil.MockHabitRepository{}
	logs := &testutil.MockHabitLogRepository{}
	svc := newProgressService(habits, logs)

	habits.On("GetByIDForUser", mock.Anything, "habit-1", "user-2").
		Return(nil, domain.ErrForbidden)

	_, err := svc.IncrementProgress(context.Background(), "user-2", "habit-1", "")
	assert.ErrorIs(t, err, domain.ErrForbidden)
	logs.AssertNotCalled(t, "IncrementProgress")
}
