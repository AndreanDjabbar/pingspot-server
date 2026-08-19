package service

import (
	"context"
	"pingspot/internal/domain/notification_service/dto"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	userRepo "pingspot/internal/domain/user_service/repository"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationService struct {
	notificationRepo notificationRepo.NotificationRepository
	userRepo           userRepo.UserRepository
	db               *gorm.DB
}

func NewNotificationService(db *gorm.DB, notificationRepo notificationRepo.NotificationRepository, userRepo userRepo.UserRepository) *NotificationService {
	return &NotificationService{
		db:               db,
		notificationRepo: notificationRepo,
		userRepo:           userRepo,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID uint) (*dto.GetNotificationsResponse, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan pengguna", err.Error(), nil)
	}
	if existingUser == nil {
		return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "Pengguna dengan ID tersebut tidak ada", nil)
	}
	notifications, err := s.notificationRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		return nil, apperror.New(500, "NOTIFICATION_FETCH_FAILED", "gagal mendapatkan notifikasi", err.Error(), nil)
	}
	var notificationsDTO []*dto.Notification
	for _, notification := range *notifications {
		notificationsDTO = append(notificationsDTO, &dto.Notification{
			ID:          notification.ID,
			UserID:      notification.UserID,
			Type:        string(notification.Type),
			Title:       notification.Title,
			Description: notification.Description,
			Category:    string(notification.Category),
			IsRead:      *notification.IsRead,
			ReadAt:      notification.ReadAt,
			CreatedAt:   notification.CreatedAt,
			EntityID:    notification.EntityID,
			EntityType:  (*string)(notification.EntityType),
		})
	}
	return &dto.GetNotificationsResponse{
		Notifications: notificationsDTO,
	}, nil
}