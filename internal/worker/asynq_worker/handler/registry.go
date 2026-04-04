package handler

import (
	"pingspot/internal/domain/report_service/repository"
	taskHandler "pingspot/internal/domain/task_service/handler"
	"pingspot/internal/domain/task_service/tasks"
	"pingspot/internal/infrastructure/database"

	"github.com/hibiken/asynq"
)

func RegisterAllHandlers(mux *asynq.ServeMux) {
	db := database.GetPostgresDB()
	reportRepo := repository.NewReportRepository(db)
	taskHandler := taskHandler.NewTaskHandler(db, reportRepo)

	mux.HandleFunc(tasks.TaskAutoResolveReport, taskHandler.AutoResolveReportHandler)
}
