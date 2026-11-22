package repository

import (
	"context"
	"pingspot/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	WithTX(tx *gorm.DB) NotificationRepository
	GetByID(ctx context.Context, id uint) (*model.Notification, error)
	GetByUserID(ctx context.Context, userID uint) (*[]model.Notification, error)
	Create(ctx context.Context, notification *model.Notification) error
	Update(ctx context.Context, notification *model.Notification) error
	MarkAllAsRead(ctx context.Context, userID uint) error
	DeleteByID(ctx context.Context, id uint) error
	DeleteByUserID(ctx context.Context, userID uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) WithTX(tx *gorm.DB) NotificationRepository {
	return &notificationRepository{
		db: tx,
	}
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uint) (*[]model.Notification, error) {
	var notifications []model.Notification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return &notifications, nil
}

func (r *notificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	if err := r.db.WithContext(ctx).Create(notification).Error; err != nil {
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

func (r *notificationRepository) Update(ctx context.Context, notification *model.Notification) error {
	if err := r.db.WithContext(ctx).Save(notification).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Updates(map[string]any{
		"is_read": true,
		"read_at": gorm.Expr("EXTRACT(EPOCH FROM NOW())::BIGINT"),
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) DeleteByID(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Notification{}, id).Error; err != nil {
		return err
	}
	return nil
}

func (r *notificationRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.Notification{}).Error; err != nil {
		return err
	}
	return nil
}