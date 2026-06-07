package postgres

import (
	"context"
	"errors"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
	"gorm.io/gorm"
)

type habitRepo struct {
	db *gorm.DB
}

// NewHabitRepository returns a PostgreSQL-backed HabitRepository.
func NewHabitRepository(db *gorm.DB) repository.HabitRepository {
	return &habitRepo{db: db}
}

func (r *habitRepo) Create(ctx context.Context, habit *domain.Habit) error {
	m := HabitModelFromDomain(habit)
	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		return result.Error
	}
	habit.ID = m.ID
	habit.CreatedAt = m.CreatedAt
	habit.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *habitRepo) GetByID(ctx context.Context, id string) (*domain.Habit, error) {
	var m HabitModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return m.ToDomain(), nil
}

func (r *habitRepo) GetByIDForUser(ctx context.Context, id, userID string) (*domain.Habit, error) {
	habit, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if habit.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return habit, nil
}

func (r *habitRepo) Update(ctx context.Context, habit *domain.Habit) error {
	m := HabitModelFromDomain(habit)
	m.ID = habit.ID
	// Use Select to update only the fields we allow mutation on.
	result := r.db.WithContext(ctx).Model(m).
		Select("title", "color", "target_per_day", "schedule_type", "weekdays", "updated_at").
		Updates(m)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *habitRepo) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&HabitModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *habitRepo) List(ctx context.Context, filter repository.HabitFilter) ([]*domain.Habit, int64, error) {
	var models []HabitModel
	var total int64

	q := r.db.WithContext(ctx).Model(&HabitModel{}).Where("user_id = ?", filter.UserID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	size := filter.Size
	if size < 1 || size > 200 {
		size = 50
	}
	offset := (page - 1) * size

	if err := q.Order("created_at DESC").Offset(offset).Limit(size).Find(&models).Error; err != nil {
		return nil, 0, err
	}

	habits := make([]*domain.Habit, len(models))
	for i, m := range models {
		m := m
		habits[i] = m.ToDomain()
	}
	return habits, total, nil
}

func (r *habitRepo) ListForUser(ctx context.Context, userID string) ([]*domain.Habit, error) {
	var models []HabitModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}
	habits := make([]*domain.Habit, len(models))
	for i, m := range models {
		m := m
		habits[i] = m.ToDomain()
	}
	return habits, nil
}
