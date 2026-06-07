package postgres

import (
	"context"
	"errors"

	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/repository"
	"gorm.io/gorm"
)

type userRepo struct {
	db *gorm.DB
}

// NewUserRepository returns a PostgreSQL-backed UserRepository.
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *domain.User) error {
	m := UserModelFromDomain(user)
	result := r.db.WithContext(ctx).Create(m)
	if result.Error != nil {
		if isUniqueViolation(result.Error) {
			return domain.ErrAlreadyExists
		}
		return result.Error
	}
	// Back-fill generated fields.
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var m UserModel
	result := r.db.WithContext(ctx).Where("id = ?", id).First(&m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return m.ToDomain(), nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m UserModel
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&m)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, result.Error
	}
	return m.ToDomain(), nil
}

func (r *userRepo) Update(ctx context.Context, user *domain.User) error {
	m := UserModelFromDomain(user)
	result := r.db.WithContext(ctx).Save(m)
	return result.Error
}
