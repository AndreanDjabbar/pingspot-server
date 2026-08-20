package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	GetByUserID(ctx context.Context, userID uint) (*[]model.Notification, error)
	CreateTX(ctx context.Context, tx *gorm.DB, notification *model.Notification) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uint) (*[]model.Notification, error) {
	var notifications []model.Notification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return &notifications, nil
}

func (r *notificationRepository) CreateTX(ctx context.Context, tx *gorm.DB, notification *model.Notification) error {
	if err := tx.WithContext(ctx).Create(notification).Error; err != nil {
		return err
	}
	return nil
}