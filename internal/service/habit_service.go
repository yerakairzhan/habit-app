package service

import (
	"context"
	"time"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
)

// HabitService manages habit CRUD and the today-view use-cases.
type HabitService struct {
	habits   repository.HabitRepository
	logs     repository.HabitLogRepository
}

func NewHabitService(
	habits repository.HabitRepository,
	logs repository.HabitLogRepository,
) *HabitService {
	return &HabitService{habits: habits, logs: logs}
}

// CreateHabit validates and persists a new habit for the authenticated user.
func (s *HabitService) CreateHabit(ctx context.Context, userID string, h *domain.Habit) (*domain.Habit, error) {
	if err := validateHabit(h); err != nil {
		return nil, err
	}
	h.UserID = userID
	if err := s.habits.Create(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

// UpdateHabit applies field-level updates. Only the requesting user may update their habits.
func (s *HabitService) UpdateHabit(ctx context.Context, userID string, h *domain.Habit) (*domain.Habit, error) {
	existing, err := s.habits.GetByIDForUser(ctx, h.ID, userID)
	if err != nil {
		return nil, err
	}

	// Apply updates onto the existing entity.
	existing.Title = h.Title
	existing.Color = h.Color
	existing.TargetPerDay = h.TargetPerDay
	existing.ScheduleType = h.ScheduleType
	existing.Weekdays = h.Weekdays

	if err := validateHabit(existing); err != nil {
		return nil, err
	}
	if err := s.habits.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// DeleteHabit soft-deletes a habit (GORM DeletedAt).
func (s *HabitService) DeleteHabit(ctx context.Context, userID, habitID string) error {
	if _, err := s.habits.GetByIDForUser(ctx, habitID, userID); err != nil {
		return err
	}
	return s.habits.Delete(ctx, habitID)
}

// GetHabit retrieves a single habit, enforcing ownership.
func (s *HabitService) GetHabit(ctx context.Context, userID, habitID string) (*domain.Habit, error) {
	return s.habits.GetByIDForUser(ctx, habitID, userID)
}

// ListHabits returns a paginated list of habits for the user.
func (s *HabitService) ListHabits(ctx context.Context, userID string, page, size int) ([]*domain.Habit, int64, error) {
	return s.habits.List(ctx, repository.HabitFilter{
		UserID: userID,
		Page:   page,
		Size:   size,
	})
}

// TodayHabitView bundles a habit with today's progress snapshot.
type TodayHabitView struct {
	Habit     *domain.Habit
	Progress  int
	Target    int
	Completed bool
}

// GetTodaysHabits returns only the habits scheduled for today, each with current progress.
func (s *HabitService) GetTodaysHabits(ctx context.Context, userID string) ([]*TodayHabitView, error) {
	today := time.Now().UTC()

	habits, err := s.habits.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var views []*TodayHabitView
	for _, h := range habits {
		if !h.IsScheduledOn(today) {
			continue
		}

		log, err := s.logs.GetByHabitAndDate(ctx, h.ID, today)
		progress := 0
		if err == nil {
			progress = log.Progress
		} else if err != domain.ErrNotFound {
			return nil, err
		}

		views = append(views, &TodayHabitView{
			Habit:     h,
			Progress:  progress,
			Target:    h.TargetPerDay,
			Completed: progress >= h.TargetPerDay,
		})
	}
	return views, nil
}

// ─── Validation ──────────────────────────────────────────────────────────────

func validateHabit(h *domain.Habit) error {
	if h.Title == "" {
		return domain.ErrInvalidInput
	}
	if h.TargetPerDay < 1 {
		return domain.ErrInvalidInput
	}
	if h.ScheduleType == "" {
		return domain.ErrInvalidInput
	}
	validSchedules := map[domain.ScheduleType]bool{
		domain.ScheduleDaily:         true,
		domain.ScheduleEveryOtherDay: true,
		domain.ScheduleWeekdays:      true,
	}
	if !validSchedules[h.ScheduleType] {
		return domain.ErrInvalidInput
	}
	if h.ScheduleType == domain.ScheduleWeekdays && len(h.Weekdays) == 0 {
		return domain.ErrInvalidInput
	}
	return nil
}
