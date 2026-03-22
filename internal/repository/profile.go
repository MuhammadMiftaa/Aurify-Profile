package repository

import (
	"context"
	"errors"

	"refina-profile/internal/types/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProfileRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Profile, error)
	Create(ctx context.Context, profile *model.Profile) error
	Update(ctx context.Context, profile *model.Profile) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Return nil, nil when not found (will create new)
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Profile, error) {
	var profile model.Profile
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("profile not found")
		}
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) Create(ctx context.Context, profile *model.Profile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *profileRepository) Update(ctx context.Context, profile *model.Profile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}

func (r *profileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Profile{}, id).Error
}
