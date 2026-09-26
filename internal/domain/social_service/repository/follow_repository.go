package repository

import (
	"context"
	"pingspot/internal/domain/social_service/dto"
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
	GetFollowersByUserIDPaginated(ctx context.Context, userID uint, limit int, cursorID *string) ([]*dto.GetConnectionsRepositoryResponse, error)
	GetFollowingByUserIDPaginated(ctx context.Context, userID uint, limit int, cursorID *string) ([]*dto.GetConnectionsRepositoryResponse, error)
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

func (r *followRepository) GetFollowersByUserIDPaginated(ctx context.Context, userID uint, limit int, cursorID *string) ([]*dto.GetConnectionsRepositoryResponse, error) {
	type row struct {
		model.User
		FollowID       uint
		FollowerUserID uint
		FollowingID    uint
		FollowingType  model.FollowingType
		FollowCreatedAt int64
	}

	var rows []row

	query := r.db.WithContext(ctx).
		Table("follows").
		Select(`users.*, 
			follows.id as follow_id, 
			follows.follower_user_id, 
			follows.following_id, 
			follows.following_type, 
			follows.created_at as follow_created_at`).
		Joins("JOIN users ON follows.follower_user_id = users.id").
		Where("follows.following_id = ? AND follows.following_type = ?", userID, model.FollowingTypeUser).
		Order("follows.id DESC").
		Limit(limit)

	if cursorID != nil && *cursorID != "" {
		query = query.Where("follows.id < ?", *cursorID)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]*dto.GetConnectionsRepositoryResponse, 0, len(rows))
	for _, rr := range rows {
		u := rr.User
		items = append(items, &dto.GetConnectionsRepositoryResponse{
			User: &u,
			Follow: &model.Follow{
				ID:             rr.FollowID,
				FollowerUserID: rr.FollowerUserID,
				FollowingID:    rr.FollowingID,
				FollowingType:  rr.FollowingType,
				CreatedAt:      rr.FollowCreatedAt,
			},
		})
	}

	return items, nil
}

func (r *followRepository) GetFollowingByUserIDPaginated(ctx context.Context, userID uint, limit int, cursorID *string) ([]*dto.GetConnectionsRepositoryResponse, error) {
	type row struct {
		model.User
		FollowID       uint
		FollowerUserID uint
		FollowingID    uint
		FollowingType  model.FollowingType
		FollowCreatedAt int64
	}

	var rows []row

	query := r.db.WithContext(ctx).
		Table("follows").
		Select(`users.*, 
			follows.id as follow_id, 
			follows.follower_user_id, 
			follows.following_id, 
			follows.following_type, 
			follows.created_at as follow_created_at`).
		Joins("JOIN users ON follows.following_id = users.id").
		Where("follows.follower_user_id = ? AND follows.following_type = ?", userID, model.FollowingTypeUser).
		Order("follows.id DESC").
		Limit(limit)

	if cursorID != nil && *cursorID != "" {
		query = query.Where("follows.id < ?", *cursorID)
	}

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	items := make([]*dto.GetConnectionsRepositoryResponse, 0, len(rows))
	for _, rr := range rows {
		u := rr.User
		items = append(items, &dto.GetConnectionsRepositoryResponse{
			User: &u,
			Follow: &model.Follow{
				ID:             rr.FollowID,
				FollowerUserID: rr.FollowerUserID,
				FollowingID:    rr.FollowingID,
				FollowingType:  rr.FollowingType,
				CreatedAt:      rr.FollowCreatedAt,
			},
		})
	}

	return items, nil
}