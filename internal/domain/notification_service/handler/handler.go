package handler

import (
	"pingspot/internal/domain/notification_service/service"
	apperror "pingspot/pkg/app_error"
	tokenutils "pingspot/pkg/utils/token_util"
	"pingspot/pkg/logger"
	response "pingspot/pkg/utils/response_util"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	notificationService *service.NotificationService
}

func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userId := uint(claims["user_id"].(float64))

	notifications, err := h.notificationService.GetNotifications(ctx, userId)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan notifikasi", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Berhasil mendapatkan notifikasi", "data", notifications)
}

func (h *NotificationHandler) MarkNotificationAsRead(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userId := uint(claims["user_id"].(float64))
	notificationID, err := c.ParamsInt("notificationID")
	if err != nil {
		logger.Error("Failed to parse notification ID", zap.Error(err))
		return response.ResponseError(c, 400, "ID notifikasi tidak valid", "", "ID notifikasi harus berupa angka")
	}
	err = h.notificationService.MarkNotificationAsRead(ctx, userId, uint(notificationID))
	if err != nil {
		logger.Error("Failed to mark notification as read", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal menandai notifikasi sebagai dibaca", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Berhasil menandai notifikasi sebagai dibaca", "data", nil)
}