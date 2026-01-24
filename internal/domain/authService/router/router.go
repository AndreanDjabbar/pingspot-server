package router

import (
	"pingspot/internal/domain/authService/handler"
	"pingspot/internal/domain/authService/service"
	userRepository "pingspot/internal/domain/userService/repository"
	"pingspot/internal/infrastructure/cache"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/middleware"
	cacheRepository "pingspot/internal/repository"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(app *fiber.App) {
	db := database.GetPostgresDB()
	rdb := cache.GetRedis()

	userRepo := userRepository.NewUserRepository(db)
	userProfileRepo := userRepository.NewUserProfileRepository(db)
	userSessionRepo := userRepository.NewUserSessionRepository(db)
	cacheRepo := cacheRepository.NewCacheRepository(&rdb)
	authService := service.NewAuthService(userRepo, userProfileRepo, userSessionRepo, cacheRepo)
	authHandler := handler.NewAuthHandler(authService)

	authRoute := app.Group("/pingspot/api/auth")

	authRoute.Post("/verification", 
	middleware.TimeoutMiddleware(10*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 6,
		KeyPrefix: "email_verification",
	})),
	authHandler.VerificationHandler,
	)

	authRoute.Post("/register", 
	middleware.TimeoutMiddleware(15*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 6,
		KeyPrefix: "register",
	})),
	authHandler.RegisterHandler,
	)

	authRoute.Post("/login", 
	middleware.TimeoutMiddleware(10*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 6,
		KeyPrefix: "login",
	})),
	authHandler.LoginHandler,
	)

	authRoute.Post("/logout", 
	middleware.TimeoutMiddleware(5*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 5,
		KeyPrefix: "logout",

	})), 
	authHandler.LogoutHandler,
	)

	authRoute.Post("/forgot-password/email-verification",
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 6,
		KeyPrefix: "forgot_password_email_verification",
	})),
	middleware.TimeoutMiddleware(10*time.Second), 
	authHandler.ForgotPasswordEmailVerificationHandler,
	)

	authRoute.Post("/forgot-password/link-verification", 
	middleware.TimeoutMiddleware(5*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 6,
		KeyPrefix: "forgot_password_link_verification",
	})),
	authHandler.ForgotPasswordLinkVerificationHandler,
	)

	authRoute.Post("/forgot-password/reset-password", 
	middleware.TimeoutMiddleware(5*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 6,
		KeyPrefix: "forgot_password_reset_password",
	})),
	authHandler.ForgotPasswordResetPasswordHandler,
	)

	authRoute.Post("/refresh-token", 
	middleware.TimeoutMiddleware(8*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 20,
		KeyPrefix: "refresh_token",
	})),
	authHandler.RefreshTokenHandler,
	)

	authRoute.Get("/google", 
	middleware.TimeoutMiddleware(3*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 10,
		KeyPrefix: "google_auth",
	})),
	adaptor.HTTPHandlerFunc(authHandler.GoogleLoginHandler),
	)

	authRoute.Get("/google/callback", 
	middleware.TimeoutMiddleware(10*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 10,
		KeyPrefix: "google_callback",
	})),
	adaptor.HTTPHandlerFunc(authHandler.GoogleCallbackHandler),
	)
}
