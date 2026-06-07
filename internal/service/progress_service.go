package service

import (
	"context"
	"time"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
)

// ProgressService manages daily progress mutations.
type ProgressService struct {
	habits repository.HabitRepository
	logs   repository.HabitLogRepository
}

func NewProgressService(
	habits repository.HabitRepository,
	logs repository.HabitLogRepository,
) *ProgressService {
	return &ProgressService{habits: habits, logs: logs}
}

// ProgressResult is returned by all progress mutations.
type ProgressResult struct {
	HabitID   string
	Date      string
	Progress  int
	Target    int
	Completed bool
}

func (s *ProgressService) IncrementProgress(ctx context.Context, userID, habitID, dateStr string) (*ProgressResult, error) {
	habit, date, err := s.resolveHabitAndDate(ctx, userID, habitID, dateStr)
	if err != nil {
		return nil, err
	}
	log, err := s.logs.IncrementProgress(ctx, habitID, userID, date)
	if err != nil {
		return nil, err
	}
	return toProgressResult(habit, log, date), nil
}

func (s *ProgressService) DecrementProgress(ctx context.Context, userID, habitID, dateStr string) (*ProgressResult, error) {
	habit, date, err := s.resolveHabitAndDate(ctx, userID, habitID, dateStr)
	if err != nil {
		return nil, err
	}
	log, err := s.logs.DecrementProgress(ctx, habitID, userID, date)
	if err != nil {
		return nil, err
	}
	return toProgressResult(habit, log, date), nil
}

func (s *ProgressService) ResetProgress(ctx context.Context, userID, habitID, dateStr string) (*ProgressResult, error) {
	habit, date, err := s.resolveHabitAndDate(ctx, userID, habitID, dateStr)
	if err != nil {
		return nil, err
	}
	log, err := s.logs.ResetProgress(ctx, habitID, userID, date)
	if err != nil {
		return nil, err
	}
	return toProgressResult(habit, log, date), nil
}

func (s *ProgressService) GetProgress(ctx context.Context, userID, habitID, dateStr string) (*ProgressResult, error) {
	habit, date, err := s.resolveHabitAndDate(ctx, userID, habitID, dateStr)
	if err != nil {
		return nil, err
	}
	log, err := s.logs.GetByHabitAndDate(ctx, habitID, date)
	if err == domain.ErrNotFound {
		// No log yet — return zero progress.
		return &ProgressResult{
			HabitID:   habitID,
			Date:      date.Format("2006-01-02"),
			Progress:  0,
			Target:    habit.TargetPerDay,
			Completed: false,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return toProgressResult(habit, log, date), nil
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// resolveHabitAndDate fetches the habit (enforcing ownership) and parses the date.
// An empty dateStr defaults to today UTC.
func (s *ProgressService) resolveHabitAndDate(ctx context.Context, userID, habitID, dateStr string) (*domain.Habit, time.Time, error) {
	habit, err := s.habits.GetByIDForUser(ctx, habitID, userID)
	if err != nil {
		return nil, time.Time{}, err
	}

	var date time.Time
	if dateStr == "" {
		date = time.Now().UTC()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return nil, time.Time{}, domain.ErrInvalidInput
		}
	}
	return habit, date, nil
}

func toProgressResult(habit *domain.Habit, log *domain.HabitLog, date time.Time) *ProgressResult {
	return &ProgressResult{
		HabitID:   habit.ID,
		Date:      date.UTC().Format("2006-01-02"),
		Progress:  log.Progress,
		Target:    habit.TargetPerDay,
		Completed: log.Completed(habit.TargetPerDay),
	}
}
