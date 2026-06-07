package service

import (
	"context"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
)

// SettingsService manages user preferences.
type SettingsService struct {
	settings repository.SettingsRepository
}

func NewSettingsService(settings repository.SettingsRepository) *SettingsService {
	return &SettingsService{settings: settings}
}

func (s *SettingsService) GetSettings(ctx context.Context, userID string) (*domain.Settings, error) {
	return s.settings.GetOrCreate(ctx, userID)
}

func (s *SettingsService) UpdateSettings(ctx context.Context, userID string, in *domain.Settings) (*domain.Settings, error) {
	existing, err := s.settings.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err
	}

	if in.Language != "" {
		existing.Language = in.Language
	}
	existing.NotificationsEnabled = in.NotificationsEnabled
	existing.ShowDates = in.ShowDates

	if err := s.settings.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}
