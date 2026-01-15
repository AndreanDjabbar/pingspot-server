package router

import (
	reportRepository "pingspot/internal/domain/reportService/repository"
	"pingspot/internal/domain/searchService/handler"
	"pingspot/internal/domain/searchService/service"
	userRepository "pingspot/internal/domain/userService/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/middleware"
	"time"

	"github.com/gofiber/fiber/v2"
)

func RegisterSearchRoutes(app *fiber.App) {
	db := database.GetPostgresDB()
	userRepo := userRepository.NewUserRepository(db)
	reportRepo := reportRepository.NewReportRepository(db)

	searchService := service.NewSearchService(userRepo, reportRepo)
	searchHandler := handler.NewSearchHandler(searchService)

	searchRoute := app.Group("/pingspot/api/search", middleware.ValidateAccessToken())
	searchRoute.Get("/", 
	middleware.TimeoutMiddleware(40*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 150,
		KeyPrefix: "search_requests",
	})), 
	searchHandler.HandleSearch,
	)
}
