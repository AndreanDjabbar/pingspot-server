package handler

import (
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	cacheRepository "pingspot/internal/repository"
	reportRepo "pingspot/internal/domain/report_service/repository"
	socialRepo "pingspot/internal/domain/social_service/repository"
	taskHandler "pingspot/internal/domain/task_service/handler"
	"pingspot/internal/domain/task_service/tasks"
	"pingspot/internal/infrastructure/cache"
	"pingspot/internal/infrastructure/database"

	"github.com/hibiken/asynq"
)

func RegisterAllHandlers(mux *asynq.ServeMux) {
	db := database.GetPostgresDB()
	rdb := cache.GetRedis()

	reportRepo := reportRepo.NewReportRepository(db)
	socialRepo := socialRepo.NewFollowRepository(db)
	notificationRepo := notificationRepo.NewNotificationRepository(db)
	cacheRepo := cacheRepository.NewCacheRepository(&rdb)

	taskHandler := taskHandler.NewTaskHandler(db, cacheRepo, reportRepo, notificationRepo, socialRepo)

	mux.HandleFunc(tasks.TaskAutoResolveReport, taskHandler.AutoResolveReportHandler)
	mux.HandleFunc(tasks.TaskCreateNotification, taskHandler.CreateNotificationHandler)
	mux.HandleFunc(tasks.TaskSendFollowerReportNotification, taskHandler.SendFollowerReportNotificationHandler)
	mux.HandleFunc(tasks.TaskSendReportCommentNotification, taskHandler.SendReportCommentNotificationHandler)
}
