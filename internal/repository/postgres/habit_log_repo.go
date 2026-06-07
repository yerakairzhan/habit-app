package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type habitLogRepo struct {
	db *gorm.DB
}

// NewHabitLogRepository returns a PostgreSQL-backed HabitLogRepository.
func NewHabitLogRepository(db *gorm.DB) repository.HabitLogRepository {
	return &habitLogRepo{db: db}
}

func (r *habitLogRepo) Upsert(ctx context.Context, log *domain.HabitLog) error {
	m := HabitLogModelFromDomain(log)
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "habit_id"}, {Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{"progress", "updated_at"}),
		}).
		Create(m)
	if result.Error != nil {
		return result.Error
	}
	log.ID = m.ID
	log.CreatedAt = m.CreatedAt
	log.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *habitLogRepo) GetByHabitAndDate(ctx context.Context, habitID string, date time.Time) (*domain.HabitLog, error) {
	var m HabitLogModel
	d := truncateToDay(date)
	result := r.db.WithContext(ctx).
		Where("habit_id = ? AND date = ?", habitID, d).
		First(&m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return m.ToDomain(), nil
}

// IncrementProgress uses a single atomic UPDATE … RETURNING to avoid race conditions.
// If no row exists yet, it inserts one with progress=1.
func (r *habitLogRepo) IncrementProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error) {
	d := truncateToDay(date)
	// return r.adjustProgress(ctx, habitID, userID, d, "GREATEST(0, progress + 1)")
	return r.adjustProgress(ctx, habitID, userID, d, "GREATEST(0, habit_logs.progress + 1)")
}

func (r *habitLogRepo) DecrementProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error) {
	d := truncateToDay(date)
	return r.adjustProgress(ctx, habitID, userID, d, "GREATEST(0, progress - 1)")
}

func (r *habitLogRepo) ResetProgress(ctx context.Context, habitID, userID string, date time.Time) (*domain.HabitLog, error) {
	d := truncateToDay(date)
	return r.adjustProgress(ctx, habitID, userID, d, "0")
}

// adjustProgress performs an INSERT … ON CONFLICT DO UPDATE atomically,
// applying the given SQL expression to the progress column.
func (r *habitLogRepo) adjustProgress(ctx context.Context, habitID, userID string, date time.Time, expr string) (*domain.HabitLog, error) {
	newID := uuid.NewString()

	// The raw SQL is intentional: GORM does not expose INSERT … ON CONFLICT UPDATE RETURNING.
	// We use parameterised values to prevent injection.
	type result struct {
		ID        string
		HabitID   string
		UserID    string
		Date      time.Time
		Progress  int
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	var row result

	// language=PostgreSQL
	sql := `
		INSERT INTO habit_logs (id, habit_id, user_id, date, progress, created_at, updated_at)
		VALUES (?, ?, ?, ?, 1, NOW(), NOW())
		ON CONFLICT (habit_id, date)
		DO UPDATE SET
			progress   = ` + expr + `,
			updated_at = NOW()
		RETURNING id, habit_id, user_id, date, progress, created_at, updated_at
	`
	if err := r.db.WithContext(ctx).Raw(sql, newID, habitID, userID, date).Scan(&row).Error; err != nil {
		return nil, err
	}

	return &domain.HabitLog{
		ID:        row.ID,
		HabitID:   row.HabitID,
		UserID:    row.UserID,
		Date:      row.Date,
		Progress:  row.Progress,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *habitLogRepo) ListByHabitSince(ctx context.Context, habitID string, since time.Time) ([]*domain.HabitLog, error) {
	var models []HabitLogModel
	if err := r.db.WithContext(ctx).
		Where("habit_id = ? AND date >= ?", habitID, truncateToDay(since)).
		Order("date ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	return toLogSlice(models), nil
}

func (r *habitLogRepo) ListByUserSince(ctx context.Context, userID string, since time.Time) ([]*domain.HabitLog, error) {
	var models []HabitLogModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND date >= ?", userID, truncateToDay(since)).
		Order("date ASC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	return toLogSlice(models), nil
}

func (r *habitLogRepo) CountTrackedHabits(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&HabitLogModel{}).
		Where("user_id = ? AND progress > 0", userID).
		Distinct("habit_id").
		Count(&count).Error
	return int(count), err
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func truncateToDay(t time.Time) time.Time {
	return t.UTC().Truncate(24 * time.Hour)
}

func toLogSlice(models []HabitLogModel) []*domain.HabitLog {
	logs := make([]*domain.HabitLog, len(models))
	for i, m := range models {
		m := m
		logs[i] = m.ToDomain()
	}
	return logs
}
