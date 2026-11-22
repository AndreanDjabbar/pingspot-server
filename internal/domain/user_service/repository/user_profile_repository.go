package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type UserProfileRepository interface {
	WithTX(tx *gorm.DB) UserProfileRepository
	SaveByID(ctx context.Context, userID uint, profile *model.UserProfile) (*model.UserProfile, error)
	GetByID(ctx context.Context, userID uint) (*model.UserProfile, error)
	Create(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error)
	Update(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error)
}

type userProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: db}
}

func (r *userProfileRepository) WithTX(tx *gorm.DB) UserProfileRepository {
	return &userProfileRepository{
		db: tx,
	}
}

func (r *userProfileRepository) SaveByID(ctx context.Context, userID uint, profile *model.UserProfile) (*model.UserProfile, error) {
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *userProfileRepository) GetByID(ctx context.Context, userID uint) (*model.UserProfile, error) {
	var profile model.UserProfile
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepository) Create(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	if err := r.db.WithContext(ctx).Create(profile).Error; err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *userProfileRepository) Update(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error) {
	if err := r.db.WithContext(ctx).Save(profile).Error; err != nil {
		return nil, err
	}
	return profile, nil
}
