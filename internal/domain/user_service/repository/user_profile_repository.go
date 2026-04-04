package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type UserProfileRepository interface {
	SaveByID(ctx context.Context, userID uint, profile *model.UserProfile) (*model.UserProfile, error)
	GetByID(ctx context.Context, userID uint) (*model.UserProfile, error)
	GetByIDTX(ctx context.Context, tx *gorm.DB, userID uint) (*model.UserProfile, error)
	CreateTX(ctx context.Context, tx *gorm.DB, profile *model.UserProfile) (*model.UserProfile, error)
	UpdateTX(ctx context.Context, tx *gorm.DB, profile *model.UserProfile) (*model.UserProfile, error)
}

type userProfileRepository struct {
	db *gorm.DB
}

func NewUserProfileRepository(db *gorm.DB) UserProfileRepository {
	return &userProfileRepository{db: db}
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

func (r *userProfileRepository) GetByIDTX(ctx context.Context, tx *gorm.DB, userID uint) (*model.UserProfile, error) {
	var profile model.UserProfile
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *userProfileRepository) CreateTX(ctx context.Context, tx *gorm.DB, profile *model.UserProfile) (*model.UserProfile, error) {
	if err := tx.WithContext(ctx).Create(profile).Error; err != nil {
		return nil, err
	}
	return profile, nil
}

func (r *userProfileRepository) UpdateTX(ctx context.Context, tx *gorm.DB, profile *model.UserProfile) (*model.UserProfile, error) {
	if err := tx.WithContext(ctx).Save(profile).Error; err != nil {
		return nil, err
	}
	return profile, nil
}
