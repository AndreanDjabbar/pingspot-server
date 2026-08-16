package handler

import (
	reportRepo "pingspot/internal/domain/report_service/repository"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	taskHandler "pingspot/internal/domain/task_service/handler"
	"pingspot/internal/domain/task_service/tasks"
	"pingspot/internal/infrastructure/database"

	"github.com/hibiken/asynq"
)

func RegisterAllHandlers(mux *asynq.ServeMux) {
	db := database.GetPostgresDB()
	reportRepo := reportRepo.NewReportRepository(db)
	notificationRepo := notificationRepo.NewNotificationRepository(db)
	taskHandler := taskHandler.NewTaskHandler(db, reportRepo, notificationRepo)

	mux.HandleFunc(tasks.TaskAutoResolveReport, taskHandler.AutoResolveReportHandler)
	mux.HandleFunc(tasks.TaskCreateNotification, taskHandler.CreateNotificationHandler)
}
