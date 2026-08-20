package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	GetByID(ctx context.Context, id uint) (*model.Notification, error)
	GetByUserID(ctx context.Context, userID uint) (*[]model.Notification, error)
	CreateTX(ctx context.Context, tx *gorm.DB, notification *model.Notification) error
	UpdateTX(ctx context.Context, tx *gorm.DB, notification *model.Notification) error
	MarkAllAsReadTX(ctx context.Context, tx *gorm.DB, userID uint) error
	DeleteByIDTX(ctx context.Context, tx *gorm.DB, id uint) error
	DeleteByUserIDTX(ctx context.Context, tx *gorm.DB, userID uint) error
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

func (r *notificationRepository) GetByID(ctx context.Context, id uint) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.WithContext(ctx).First(&notification, id).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) UpdateTX(ctx context.Context, tx *gorm.DB, notification *model.Notification) error {
	
	if err := tx.WithContext(ctx).Save(notification).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) MarkAllAsReadTX(ctx context.Context, tx *gorm.DB, userID uint) error {
	if err := tx.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]any{
		"is_read": true,
		"read_at": gorm.Expr("EXTRACT(EPOCH FROM NOW())::BIGINT"),
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) DeleteByIDTX(ctx context.Context, tx *gorm.DB, id uint) error {
	if err := tx.WithContext(ctx).Delete(&model.Notification{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) DeleteByUserIDTX(ctx context.Context, tx *gorm.DB, userID uint) error {
	if err := tx.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Notification{}).Error; err != nil {
		return err
	}
	return nil
}