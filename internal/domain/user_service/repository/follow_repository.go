package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type FollowRepository interface {
	CreateTX(ctx context.Context, tx *gorm.DB, follow *model.Follow) error
}

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) CreateTX(ctx context.Context, tx *gorm.DB, follow *model.Follow) error {
	if err := tx.WithContext(ctx).Create(follow).Error; err != nil {
		return err
	}
	return nil
}