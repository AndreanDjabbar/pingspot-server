package router

import (
	"pingspot/internal/domain/userService/handler"
	"pingspot/internal/domain/userService/repository"
	"pingspot/internal/domain/userService/service"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	db := database.GetPostgresDB()
	userRepo := repository.NewUserRepository(db)
	userProfileRepo := repository.NewUserProfileRepository(db)
	userService := service.NewUserService(userRepo, userProfileRepo)
	userHandler := handler.NewUserHandler(userService)

	userRoute := app.Group("/pingspot/api/user", middleware.ValidateAccessToken())

	userRoute.Get("/statistics", 
	middleware.TimeoutMiddleware(15*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "user_statistics",
	})),  
	userHandler.GetUserStatistics,
	)

	profileRoute := app.Group("/pingspot/api/user/profile", middleware.ValidateAccessToken())

	profileRoute.Get("/", 
	middleware.TimeoutMiddleware(5*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 50,
		KeyPrefix: "get_user_profile",
	})), 
	userHandler.GetProfileHandler,
	)
	profileRoute.Get("/:username", 
	middleware.TimeoutMiddleware(5*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "get_profile_by_username",
	})), 
	userHandler.GetProfileByUsernameHandler,
	)
	profileRoute.Post("/", 
	middleware.TimeoutMiddleware(10*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "save_user_profile",
	})), 
	userHandler.SaveUserProfileHandler,
	)

	securityRoute := app.Group("/pingspot/api/user/security", middleware.ValidateAccessToken())
	
	securityRoute.Post("/", 
	middleware.TimeoutMiddleware(10*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 50,
		KeyPrefix: "save_user_security",
	})),  
	userHandler.SaveUserSecurityHandler,
	)
}
