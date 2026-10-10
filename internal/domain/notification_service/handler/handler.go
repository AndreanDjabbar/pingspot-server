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
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userId := uint(claims["user_id"].(float64))

	notifications, err := h.notificationService.GetNotifications(ctx, userId)
	if err != nil {
		logger.Error("Failed to get notifications", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve notifications", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Notifications retrieved successfully", "data", notifications)
}

func (h *NotificationHandler) MarkNotificationAsRead(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userId := uint(claims["user_id"].(float64))
	notificationID, err := c.ParamsInt("notificationID")
	if err != nil {
		logger.Error("Failed to parse notification ID", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid notification ID", "", "Notification ID must be a number")
	}
	err = h.notificationService.MarkNotificationAsRead(ctx, userId, uint(notificationID))
	if err != nil {
		logger.Error("Failed to mark notification as read", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to mark notification as read", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Notification marked as read successfully", "data", nil)
}

func (h *NotificationHandler) MarkAllNotificationsAsRead(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userId := uint(claims["user_id"].(float64))
	err = h.notificationService.MarkAllNotificationsAsRead(ctx, userId)
	if err != nil {
		logger.Error("Failed to mark all notifications as read", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to mark all notifications as read", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "All notifications marked as read successfully", "data", nil)
}

func (h *NotificationHandler) DeleteNotification(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userId := uint(claims["user_id"].(float64))
	notificationID, err := c.ParamsInt("notificationID")
	if err != nil {
		logger.Error("Failed to parse notification ID", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid notification ID", "", "Notification ID must be a number")
	}
	err = h.notificationService.DeleteNotification(ctx, userId, uint(notificationID))
	if err != nil {
		logger.Error("Failed to delete notification", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to delete notification", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Notification deleted successfully", "data", nil)
}

func (h *NotificationHandler) DeleteAllNotifications(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}

	userId := uint(claims["user_id"].(float64))
	err = h.notificationService.DeleteAllNotifications(ctx, userId)
	if err != nil {
		logger.Error("Failed to delete all notifications", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to delete all notifications", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "All notifications deleted successfully", "data", nil)
}