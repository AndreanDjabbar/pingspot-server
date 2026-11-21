package service

import (
	"context"
	"pingspot/internal/domain/notification_service/dto"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	userRepo "pingspot/internal/domain/user_service/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/logger"
	"pingspot/internal/util"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NotificationService struct {
	notificationRepo notificationRepo.NotificationRepository
	userRepo           userRepo.UserRepository
	txManager          database.TransactionManager
}

func NewNotificationService(txManager database.TransactionManager, notificationRepo notificationRepo.NotificationRepository, userRepo userRepo.UserRepository) *NotificationService {
	return &NotificationService{
		txManager:          txManager,
		notificationRepo: notificationRepo,
		userRepo:           userRepo,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID uint) (*dto.GetNotificationsResponse, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user", err.Error(), nil)
	}
	if existingUser == nil {
		return nil, util.New(404, "USER_NOT_FOUND", "User not found", "No user exists with that ID", nil)
	}
	notifications, err := s.notificationRepo.GetByUserID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		return nil, util.New(500, "NOTIFICATION_FETCH_FAILED", "Failed to retrieve notifications", err.Error(), nil)
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

func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, userID uint, notificationID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user", err.Error(), nil)
	}
	if existingUser == nil {
		return util.New(404, "USER_NOT_FOUND", "User not found", "No user exists with that ID", nil)
	}
	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		return util.New(500, "NOTIFICATION_FETCH_FAILED", "Failed to retrieve notifications", err.Error(), nil)
	}

	if notification == nil {
		return util.New(404, "NOTIFICATION_NOT_FOUND", "Notification not found", "The notification with this ID does not belong to this user", nil)
	}

	if notification.UserID != userID {
		return util.New(403, "NOTIFICATION_FORBIDDEN", "Notification does not belong to this user", "You do not have permission to mark this notification as read", nil)
	}
	
	notification.IsRead = util.BoolPtrOrNil(true)
	notification.ReadAt = util.Int64PtrOrNil(time.Now().Unix())

	err = s.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.notificationRepo.UpdateTX(ctx, tx, notification); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to update notification", zap.Error(err))
		return util.New(500, "NOTIFICATION_UPDATE_FAILED", "Failed to update notification", err.Error(), nil)
	}
	return nil
}

func (s *NotificationService) MarkAllNotificationsAsRead(ctx context.Context, userID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user", err.Error(), nil)
	}
	if existingUser == nil {
		return util.New(404, "USER_NOT_FOUND", "User not found", "No user exists with that ID", nil)
	}

	err = s.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.notificationRepo.MarkAllAsReadTX(ctx, tx, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to mark all notifications as read", zap.Error(err))
		return util.New(500, "NOTIFICATION_MARK_ALL_AS_READ_FAILED", "Failed to mark all notifications as read", err.Error(), nil)
	}
	return nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, userID uint, notificationID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user", err.Error(), nil)
	}
	if existingUser == nil {
		return util.New(404, "USER_NOT_FOUND", "User not found", "No user exists with that ID", nil)
	}

	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		return util.New(500, "NOTIFICATION_FETCH_FAILED", "Failed to retrieve notifications", err.Error(), nil)
	}
	if notification == nil {
		return util.New(404, "NOTIFICATION_NOT_FOUND", "Notification not found", "The notification with this ID does not belong to this user", nil)
	}
	if notification.UserID != userID {
		return util.New(403, "NOTIFICATION_FORBIDDEN", "Notification does not belong to this user", "You do not have permission to delete this notification", nil)
	}

	err = s.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.notificationRepo.DeleteByIDTX(ctx, tx, notificationID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to delete notification", zap.Error(err))
		return util.New(500, "NOTIFICATION_DELETE_FAILED", "Failed to delete notification", err.Error(), nil)
	}
	return nil
}

func (s *NotificationService) DeleteAllNotifications(ctx context.Context, userID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user", err.Error(), nil)
	}
	if existingUser == nil {
		return util.New(404, "USER_NOT_FOUND", "User not found", "No user exists with that ID", nil)
	}

	err = s.txManager.Transaction(ctx, func(tx *gorm.DB) error {
		if err := s.notificationRepo.DeleteByUserIDTX(ctx, tx, userID); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		logger.Error("Failed to delete all notifications", zap.Error(err))
		return util.New(500, "NOTIFICATION_DELETE_ALL_FAILED", "Failed to delete all notifications", err.Error(), nil)
	}
	return nil
}