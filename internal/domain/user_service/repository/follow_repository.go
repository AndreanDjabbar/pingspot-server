package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type FollowRepository interface {
	CreateTX(ctx context.Context, tx *gorm.DB, follow *model.Follow) error
	GetByFollowerAndFollowing(ctx context.Context, followerUserID uint, followingID uint, followingType model.FollowingType) (*model.Follow, error)
	DeleteTX(ctx context.Context, tx *gorm.DB, follow *model.Follow) error
	GetFollowersCount(ctx context.Context, followingID uint, followingType model.FollowingType) (int64, error)
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

func (r *followRepository) GetByFollowerAndFollowing(ctx context.Context, followerUserID uint, followingID uint, followingType model.FollowingType) (*model.Follow, error) {
	var follow model.Follow
	if err := r.db.WithContext(ctx).Where("follower_user_id = ? AND following_id = ? AND following_type = ?", followerUserID, followingID, followingType).First(&follow).Error; err != nil {
		return nil, err
	}
	return &follow, nil
}

func (r *followRepository) DeleteTX(ctx context.Context, tx *gorm.DB, follow *model.Follow) error {
	if err := tx.WithContext(ctx).Delete(follow).Error; err != nil {
		return err
	}
	return nil
}

func (r *followRepository) GetFollowersCount(ctx context.Context, followingID uint, followingType model.FollowingType) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("following_id = ? AND following_type = ?", followingID, followingType).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}