package service

import (
	"context"
	"pingspot/internal/domain/notification_service/dto"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	userRepo "pingspot/internal/domain/user_service/repository"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/main_util"
	"time"

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

func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, userID uint, notificationID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan pengguna", err.Error(), nil)
	}
	if existingUser == nil {
		return apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "Pengguna dengan ID tersebut tidak ada", nil)
	}
	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		return apperror.New(500, "NOTIFICATION_FETCH_FAILED", "gagal mendapatkan notifikasi", err.Error(), nil)
	}

	if notification == nil {
		return apperror.New(404, "NOTIFICATION_NOT_FOUND", "notifikasi tidak ditemukan", "Notifikasi dengan ID tersebut tidak ada untuk pengguna ini", nil)
	}

	if notification.UserID != userID {
		return apperror.New(403, "NOTIFICATION_FORBIDDEN", "notifikasi tidak untuk pengguna ini", "Anda tidak memiliki izin untuk menandai notifikasi ini sebagai dibaca", nil)
	}
	
	notification.IsRead = main_util.BoolPtrOrNil(true)
	notification.ReadAt = main_util.Int64PtrOrNil(time.Now().Unix())

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err = s.notificationRepo.UpdateTX(ctx, tx, notification)
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to update notification", zap.Error(err))
		return apperror.New(500, "NOTIFICATION_UPDATE_FAILED", "gagal memperbarui notifikasi", err.Error(), nil)
	}
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyelesaikan transaksi", err.Error(), nil)
	}
	return nil
}

func (s *NotificationService) MarkAllNotificationsAsRead(ctx context.Context, userID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan pengguna", err.Error(), nil)
	}
	if existingUser == nil {
		return apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "Pengguna dengan ID tersebut tidak ada", nil)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	err = s.notificationRepo.MarkAllAsReadTX(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to mark all notifications as read", zap.Error(err))
		return apperror.New(500, "NOTIFICATION_MARK_ALL_AS_READ_FAILED", "gagal menandai semua notifikasi sebagai dibaca", err.Error(), nil)
	}
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyelesaikan transaksi", err.Error(), nil)
	}
	return nil
}

func (s *NotificationService) DeleteNotification(ctx context.Context, userID uint, notificationID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan pengguna", err.Error(), nil)
	}
	if existingUser == nil {
		return apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "Pengguna dengan ID tersebut tidak ada", nil)
	}

	notification, err := s.notificationRepo.GetByID(ctx, notificationID)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		return apperror.New(500, "NOTIFICATION_FETCH_FAILED", "gagal mendapatkan notifikasi", err.Error(), nil)
	}
	if notification == nil {
		return apperror.New(404, "NOTIFICATION_NOT_FOUND", "notifikasi tidak ditemukan", "Notifikasi dengan ID tersebut tidak ada untuk pengguna ini", nil)
	}
	if notification.UserID != userID {
		return apperror.New(403, "NOTIFICATION_FORBIDDEN", "notifikasi tidak untuk pengguna ini", "Anda tidak memiliki izin untuk menghapus notifikasi ini", nil)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = s.notificationRepo.DeleteByIDTX(ctx, tx, notificationID)
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to delete notification", zap.Error(err))
		return apperror.New(500, "NOTIFICATION_DELETE_FAILED", "gagal menghapus notifikasi", err.Error(), nil)
	}
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyelesaikan transaksi", err.Error(), nil)
	}
	return nil
}

func (s *NotificationService) DeleteAllNotifications(ctx context.Context, userID uint) error {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan pengguna", err.Error(), nil)
	}
	if existingUser == nil {
		return apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "Pengguna dengan ID tersebut tidak ada", nil)
	}

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err = s.notificationRepo.DeleteByUserIDTX(ctx, tx, userID)
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to delete all notifications", zap.Error(err))
		return apperror.New(500, "NOTIFICATION_DELETE_ALL_FAILED", "gagal menghapus semua notifikasi", err.Error(), nil)
	}
	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction", zap.Error(err))
		return apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyelesaikan transaksi", err.Error(), nil)
	}
	return nil
}