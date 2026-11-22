package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type UserSessionRepository interface {
	WithTX(tx *gorm.DB) UserSessionRepository
	Create(ctx context.Context, user *model.UserSession) (*model.UserSession, error)
	GetByID(ctx context.Context, id uint) (*model.UserSession, error)
	GetByRefreshTokenID(ctx context.Context, refreshTokenID string) (*model.UserSession, error)
	Update(ctx context.Context, userSession *model.UserSession) error
}

type userSessionRepository struct {
	db *gorm.DB
}

func NewUserSessionRepository(db *gorm.DB) UserSessionRepository {
	return &userSessionRepository{db: db}
}

func (r *userSessionRepository) WithTX(tx *gorm.DB) UserSessionRepository {
	return &userSessionRepository{
		db: tx,
	}
}

func (r *userSessionRepository) Create(ctx context.Context, userSession *model.UserSession) (*model.UserSession, error) {
	if err := r.db.WithContext(ctx).Create(userSession).Error; err != nil {
		return nil, err
	}
	return userSession, nil
}

func (r *userSessionRepository) GetByID(ctx context.Context, id uint) (*model.UserSession, error) {
	var userSession model.UserSession
	if err := r.db.WithContext(ctx).First(&userSession, id).Error; err != nil {
		return nil, err
	}
	return &userSession, nil
}

func (r *userSessionRepository) Update(ctx context.Context, userSession *model.UserSession) error {
	if err := r.db.WithContext(ctx).Save(userSession).Error; err != nil {
		return err
	}
	return nil
}

func (r *userSessionRepository) GetByRefreshTokenID(ctx context.Context, refreshTokenID string) (*model.UserSession, error) {
	var userSession model.UserSession
	if err := r.db.WithContext(ctx).Where("refresh_token_id = ?", refreshTokenID).First(&userSession).Error; err != nil {
		return nil, err
	}
	return &userSession, nil
}
