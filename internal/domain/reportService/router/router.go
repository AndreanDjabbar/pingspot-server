package router

import (
	"fmt"
	"pingspot/internal/domain/reportService/handler"
	reportRepository "pingspot/internal/domain/reportService/repository"
	reportService "pingspot/internal/domain/reportService/service"
	"pingspot/internal/domain/taskService/service"
	userRepository "pingspot/internal/domain/userService/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/middleware"
	"pingspot/pkg/utils/env"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"
)

func RegisterReportRoutes(app *fiber.App) {
	postgreDB := database.GetPostgresDB()
	mongoDB := database.GetMongoDB()

	reportRepo := reportRepository.NewReportRepository(postgreDB)
	reportLocationRepo := reportRepository.NewReportLocationRepository(postgreDB)
	reportImageRepo := reportRepository.NewReportImageRepository(postgreDB)
	reportReactionRepo := reportRepository.NewReportReactionRepository(postgreDB)
	reportVoteRepo := reportRepository.NewReportVoteRepository(postgreDB)
	reportProgressRepo := reportRepository.NewReportProgressRepository(postgreDB)
	userProfileRepo := userRepository.NewUserProfileRepository(postgreDB)
	userRepo := userRepository.NewUserRepository(postgreDB)
	reportCommentRepository := reportRepository.NewReportCommentRepository(mongoDB)

	redisAddress := fmt.Sprintf("%s:%s", env.RedisHost(), env.RedisPort())
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddress})
	tasksService := service.NewTaskService(client, reportRepo)

	reportService := reportService.NewreportService(reportRepo, reportLocationRepo, reportReactionRepo, reportImageRepo, userRepo, userProfileRepo, reportProgressRepo, reportVoteRepo, tasksService, reportCommentRepository)

	reportHandler := handler.NewReportHandler(reportService)

	reportRoute := app.Group("/pingspot/api/report", middleware.ValidateAccessToken())

	reportRoute.Post("/", 
	middleware.TimeoutMiddleware(20*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 8,
		KeyPrefix: "create_report",
	})), 
	reportHandler.CreateReportHandler,
	)

	reportRoute.Put("/:reportID", 
	middleware.TimeoutMiddleware(20*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      10 * time.Minute,
		MaxRequests: 15,
		KeyPrefix: "edit_report",
	})), 
	reportHandler.EditReportHandler,
	)

	reportRoute.Get("/", 
	middleware.TimeoutMiddleware(15*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "get_reports",
	})),  
	reportHandler.GetReportHandler,
	)

	reportRoute.Post("/:reportID/reaction", middleware.TimeoutMiddleware(5*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 30,
		KeyPrefix: "create_report_reaction",
	})), 
	reportHandler.ReactionReportHandler,
	)

	reportRoute.Post("/:reportID/vote", 
	middleware.TimeoutMiddleware(5*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 30,
		KeyPrefix: "create_report_vote",
	})), 
	reportHandler.VoteReportHandler,
	)

	reportRoute.Post("/:reportID/progress", middleware.TimeoutMiddleware(15*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 30,
		KeyPrefix: "upload_progress_report",
	})), 
	reportHandler.UploadProgressReportHandler,
	)

	reportRoute.Get("/:reportID/progress", middleware.TimeoutMiddleware(10*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "get_progress_report",
	})),  
	reportHandler.GetProgressReportHandler,
	)

	reportRoute.Delete("/:reportID", 
	middleware.TimeoutMiddleware(10*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 30,
		KeyPrefix: "delete_report",
	})), 
	reportHandler.DeleteReportHandler,
	)

	reportRoute.Post("/:reportID/comment", middleware.TimeoutMiddleware(8*time.Second), 
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 50,
		KeyPrefix: "create_report_comment",
	})), 
	reportHandler.CreateReportCommentHandler,
	)

	reportRoute.Get("/:reportID/comment", 
	middleware.TimeoutMiddleware(15*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "get_report_comments",
	})),  
	reportHandler.GetReportCommentsHandler,
	)

	reportRoute.Get("/comment/replies/:commentID", middleware.TimeoutMiddleware(15*time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "get_report_comment_replies",
	})),  
	reportHandler.GetReportCommentRepliesHandler,
	)

	reportRoute.Get("/statistics", 
	middleware.TimeoutMiddleware(15 * time.Second),
	middleware.UserRateLimiterMiddleware(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Window:      1 * time.Minute,
		MaxRequests: 100,
		KeyPrefix: "report_statistics",
	})),  
	reportHandler.GetReportStatisticsHandler,
	)
}
