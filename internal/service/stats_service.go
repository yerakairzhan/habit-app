package service

import (
	"context"
	"sort"
	"time"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
)

// StatsService computes habit statistics on-read.
type StatsService struct {
	habits repository.HabitRepository
	logs   repository.HabitLogRepository
}

func NewStatsService(
	habits repository.HabitRepository,
	logs repository.HabitLogRepository,
) *StatsService {
	return &StatsService{habits: habits, logs: logs}
}

// GetStats returns aggregate statistics for a user, or per-habit if habitID is set.
func (s *StatsService) GetStats(ctx context.Context, userID, habitID string) (*domain.Stats, error) {
	if habitID != "" {
		return s.habitStats(ctx, userID, habitID)
	}
	return s.userStats(ctx, userID)
}

// ─── Per-habit stats ─────────────────────────────────────────────────────────

func (s *StatsService) habitStats(ctx context.Context, userID, habitID string) (*domain.Stats, error) {
	habit, err := s.habits.GetByIDForUser(ctx, habitID, userID)
	if err != nil {
		return nil, err
	}

	// Look back up to 365 days for streak accuracy.
	since := time.Now().UTC().AddDate(-1, 0, 0)
	logs, err := s.logs.ListByHabitSince(ctx, habitID, since)
	if err != nil {
		return nil, err
	}

	logMap := toLogMap(logs)
	today := time.Now().UTC().Truncate(24 * time.Hour)

	current, best := computeStreaks(habit, logMap, today)

	tracked, err := s.logs.CountTrackedHabits(ctx, userID)
	if err != nil {
		return nil, err
	}

	rate := completionRate30d(habit, logMap, today)

	return &domain.Stats{
		CurrentStreak:      current,
		BestStreak:         best,
		TrackedHabitsCount: tracked,
		CompletionRate30d:  rate,
	}, nil
}

// ─── User-level aggregate stats ──────────────────────────────────────────────

func (s *StatsService) userStats(ctx context.Context, userID string) (*domain.Stats, error) {
	habits, err := s.habits.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	since := time.Now().UTC().AddDate(-1, 0, 0)
	logs, err := s.logs.ListByUserSince(ctx, userID, since)
	if err != nil {
		return nil, err
	}

	// Group logs by habitID.
	logsByHabit := make(map[string][]*domain.HabitLog)
	for _, l := range logs {
		logsByHabit[l.HabitID] = append(logsByHabit[l.HabitID], l)
	}

	today := time.Now().UTC().Truncate(24 * time.Hour)
	maxCurrent, maxBest := 0, 0
	var totalCompletable, totalCompleted int

	for _, h := range habits {
		logMap := toLogMap(logsByHabit[h.ID])
		current, best := computeStreaks(h, logMap, today)
		if current > maxCurrent {
			maxCurrent = current
		}
		if best > maxBest {
			maxBest = best
		}
		c, t := completionCounts30d(h, logMap, today)
		totalCompleted += c
		totalCompletable += t
	}

	tracked, err := s.logs.CountTrackedHabits(ctx, userID)
	if err != nil {
		return nil, err
	}

	var rate float32
	if totalCompletable > 0 {
		rate = float32(totalCompleted) / float32(totalCompletable)
	}

	return &domain.Stats{
		CurrentStreak:      maxCurrent,
		BestStreak:         maxBest,
		TrackedHabitsCount: tracked,
		CompletionRate30d:  rate,
	}, nil
}

// ─── Streak algorithm ────────────────────────────────────────────────────────

// computeStreaks walks backwards from today, computing current and best streaks.
// Business rule: missing a *scheduled* day breaks the streak.
// Non-scheduled days are transparent (neither extend nor break).
func computeStreaks(habit *domain.Habit, logMap map[string]*domain.HabitLog, today time.Time) (current, best int) {
	// Collect all scheduled days from creation date to today.
	start := habit.CreatedAt.UTC().Truncate(24 * time.Hour)
	var scheduled []time.Time
	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		if habit.IsScheduledOn(d) {
			scheduled = append(scheduled, d)
		}
	}

	if len(scheduled) == 0 {
		return 0, 0
	}

	// Sort descending (newest first) for current streak computation.
	sort.Slice(scheduled, func(i, j int) bool { return scheduled[i].After(scheduled[j]) })

	streak := 0
	currentActive := true

	for _, d := range scheduled {
		key := d.Format("2006-01-02")
		log, exists := logMap[key]
		completed := exists && log != nil && log.Progress >= habit.TargetPerDay

		if completed {
			streak++
			if streak > best {
				best = streak
			}
		} else {
			if currentActive {
				// For today, an incomplete day doesn't break the streak yet
				// (the user still has time). Only past incomplete days break it.
				if d.Equal(today) {
					// Don't break, but don't count today until done.
					continue
				}
				current = streak
				currentActive = false
			}
			streak = 0
		}
	}
	if currentActive {
		current = streak
	}
	return current, best
}

// ─── Completion rate ─────────────────────────────────────────────────────────

func completionRate30d(habit *domain.Habit, logMap map[string]*domain.HabitLog, today time.Time) float32 {
	completed, total := completionCounts30d(habit, logMap, today)
	if total == 0 {
		return 0
	}
	return float32(completed) / float32(total)
}

func completionCounts30d(habit *domain.Habit, logMap map[string]*domain.HabitLog, today time.Time) (completed, total int) {
	for i := 0; i < 30; i++ {
		d := today.AddDate(0, 0, -i)
		if !habit.IsScheduledOn(d) {
			continue
		}
		total++
		key := d.Format("2006-01-02")
		if log, ok := logMap[key]; ok && log.Progress >= habit.TargetPerDay {
			completed++
		}
	}
	return
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// toLogMap indexes a slice of HabitLogs by their date string (YYYY-MM-DD).
func toLogMap(logs []*domain.HabitLog) map[string]*domain.HabitLog {
	m := make(map[string]*domain.HabitLog, len(logs))
	for _, l := range logs {
		m[l.Date.UTC().Format("2006-01-02")] = l
	}
	return m
}
