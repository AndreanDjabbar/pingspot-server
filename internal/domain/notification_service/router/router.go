package router

import (
	"pingspot/internal/domain/notification_service/handler"
	userRepo "pingspot/internal/domain/user_service/repository"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	"pingspot/internal/domain/notification_service/service"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RegisterNotificationRoutes(app *fiber.App) {
	db := database.GetPostgresDB()
	userRepo := userRepo.NewUserRepository(db)
	notificationRepo := notificationRepo.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(db, notificationRepo, userRepo)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	notificationRoute := app.Group("/pingspot/api/notification", middleware.ValidateAccessToken())

	notificationRoute.Get(
		"/",
		middleware.TimeoutMiddleware(15*time.Second),
		middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
			Window:      1 * time.Minute,
			MaxRequests: 100,
			KeyPrefix: "get_notifications",
		})),
		notificationHandler.GetNotifications,
	)

	notificationRoute.Patch(
		"/read",
		middleware.TimeoutMiddleware(10 *time.Second),
		middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
			Window:      1 * time.Minute,
			MaxRequests: 50,
			KeyPrefix: "mark_all_notifications_read",
		})),
		notificationHandler.MarkAllNotificationsAsRead,
	)

	notificationRoute.Patch(
		"/:notificationID/read",
		middleware.TimeoutMiddleware(10*time.Second),
		middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
			Window:      1 * time.Minute,
			MaxRequests: 50,
			KeyPrefix: "mark_notification_read",
		})),
		notificationHandler.MarkNotificationAsRead,
	)

	notificationRoute.Delete(
		"/:notificationID",
		middleware.TimeoutMiddleware(10*time.Second),
		middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
			Window:      1 * time.Minute,
			MaxRequests: 50,
			KeyPrefix: "delete_notification",
		})),
		notificationHandler.DeleteNotification,
	)

	notificationRoute.Delete(
		"/",
		middleware.TimeoutMiddleware(10*time.Second),
		middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
			Window:      1 * time.Minute,
			MaxRequests: 50,
			KeyPrefix: "delete_all_notifications",
		})),
		notificationHandler.DeleteAllNotifications,
	)
}
