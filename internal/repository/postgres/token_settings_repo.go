package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ─── RefreshToken Repo ───────────────────────────────────────────────────────

type refreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(ctx context.Context, rt *domain.RefreshToken) error {
	m := RefreshTokenModelFromDomain(rt)
	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		return result.Error
	}
	rt.ID = m.ID
	rt.CreatedAt = m.CreatedAt
	return nil
}

func (r *refreshTokenRepo) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var m RefreshTokenModel
	result := r.db.WithContext(ctx).Where("token = ?", token).First(&m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return m.ToDomain(), nil
}

func (r *refreshTokenRepo) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token = ?", token).Delete(&RefreshTokenModel{}).Error
}

func (r *refreshTokenRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&RefreshTokenModel{}).Error
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now().UTC()).
		Delete(&RefreshTokenModel{}).Error
}

// ─── Settings Repo ───────────────────────────────────────────────────────────

type settingsRepo struct {
	db *gorm.DB
}

func NewSettingsRepository(db *gorm.DB) repository.SettingsRepository {
	return &settingsRepo{db: db}
}

func (r *settingsRepo) GetOrCreate(ctx context.Context, userID string) (*domain.Settings, error) {
	var m SettingsModel
	// Use INSERT … ON CONFLICT DO NOTHING, then SELECT.
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}},
			DoNothing: true,
		}).
		FirstOrCreate(&m, SettingsModel{UserID: userID})
	if result.Error != nil {
		return nil, result.Error
	}
	return m.ToDomain(), nil
}

func (r *settingsRepo) Update(ctx context.Context, s *domain.Settings) error {
	result := r.db.WithContext(ctx).
		Model(&SettingsModel{}).
		Where("user_id = ?", s.UserID).
		Updates(map[string]interface{}{
			"language":              s.Language,
			"notifications_enabled": s.NotificationsEnabled,
			"show_dates":            s.ShowDates,
			"updated_at":            time.Now().UTC(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}
