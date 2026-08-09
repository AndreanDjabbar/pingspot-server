package router

import (
	"pingspot/internal/domain/social_service/handler"
	"pingspot/internal/domain/social_service/service"
	userRepository "pingspot/internal/domain/user_service/repository"
	socialRepository "pingspot/internal/domain/social_service/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RegisterSocialRoutes(app *fiber.App) {
	db := database.GetPostgresDB()
	userRepo := userRepository.NewUserRepository(db)
	followRepo := socialRepository.NewFollowRepository(db)
	
	socialService := service.NewSocialService(db, followRepo, userRepo)
	socialHandler := handler.NewSocialHandler(socialService)

	followRoute := app.Group("/pingspot/api/social/follow", middleware.ValidateAccessToken())

	followRoute.Post("/", 
	middleware.TimeoutMiddleware(10*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 50,
		KeyPrefix: "follow",
	})),  
	socialHandler.FollowHandler,
	)

	followRoute.Get(
	"/:followingID/:followingType",
	middleware.TimeoutMiddleware(5*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 50,
		KeyPrefix: "get_follow_data",
	})),  
	socialHandler.GetFollowDataHandler,
	)

	connectionRoute := app.Group("/pingspot/api/social/connection", middleware.ValidateAccessToken())

	connectionRoute.Get(
	"/:userID/user",
	middleware.TimeoutMiddleware(5*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 50,
		KeyPrefix: "get_user_connections",
	})),  
	socialHandler.GetUserConnectionsHandler,
	)
}
