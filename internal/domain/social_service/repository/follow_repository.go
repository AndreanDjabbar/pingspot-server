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
	GetFollowingCount(ctx context.Context, followerUserID uint, followingType model.FollowingType) (int64, error)
	GetFollowersByUserID(ctx context.Context, userID uint) ([]*model.User, error)
	GetFollowingByUserID(ctx context.Context, userID uint) ([]*model.User, error)
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

func (r *followRepository) GetFollowingCount(ctx context.Context, followerUserID uint, followingType model.FollowingType) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Follow{}).Where("follower_user_id = ? AND following_type = ?", followerUserID, followingType).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *followRepository) GetFollowersByUserID(ctx context.Context, userID uint) ([]*model.User, error) {
	var followers []*model.User

    if err := r.db.WithContext(ctx).
        Model(&model.User{}).
        Preload("Profile").
        Joins("JOIN follows ON follows.follower_user_id = users.id").
        Where("follows.following_id = ? AND follows.following_type = ?", userID, model.FollowingTypeUser).
        Find(&followers).Error; err != nil {
        return nil, err
    }

    return followers, nil
}

func (r *followRepository) GetFollowingByUserID(ctx context.Context, userID uint) ([]*model.User, error) {
	var following []*model.User

    if err := r.db.WithContext(ctx).
        Model(&model.User{}).
        Preload("Profile").
        Joins("JOIN follows ON follows.following_id = users.id").
        Where("follows.follower_user_id = ? AND follows.following_type = ?", userID, model.FollowingTypeUser).
        Find(&following).Error; err != nil {
        return nil, err
    }

    return following, nil
}