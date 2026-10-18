package asynq_worker

import (
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	reportRepo "pingspot/internal/domain/report_service/repository"
	socialRepo "pingspot/internal/domain/social_service/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/infrastructure/redis"
	"pingspot/internal/task"

	"github.com/hibiken/asynq"
)

func RegisterAllHandlers(mux *asynq.ServeMux) {
	db := database.GetPostgresDB()
	rdb := redis.GetRedis()

	reportRepo := reportRepo.NewReportRepository(db)
	socialRepo := socialRepo.NewFollowRepository(db)
	notificationRepo := notificationRepo.NewNotificationRepository(db)
	cacheRepo := redis.NewCacheRepository(&rdb)

	taskHandler := newTaskHandler(db, cacheRepo, reportRepo, notificationRepo, socialRepo)

	mux.HandleFunc(task.TaskAutoResolveReport, taskHandler.autoResolveReportHandler)
	mux.HandleFunc(task.TaskCreateNotification, taskHandler.CreateNotificationHandler)
	mux.HandleFunc(task.TaskSendFollowerReportNotification, taskHandler.SendFollowerReportNotificationHandler)
	mux.HandleFunc(task.TaskSendReportCommentNotification, taskHandler.SendReportCommentNotificationHandler)
}
