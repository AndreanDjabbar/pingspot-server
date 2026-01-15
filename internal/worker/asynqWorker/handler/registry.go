package handler

import (
	"pingspot/internal/domain/reportService/repository"
	taskHandler "pingspot/internal/domain/taskService/handler"
	"pingspot/internal/domain/taskService/tasks"
	"pingspot/internal/infrastructure/database"

	"github.com/hibiken/asynq"
)

func RegisterAllHandlers(mux *asynq.ServeMux) {
	db := database.GetPostgresDB()
	reportRepo := repository.NewReportRepository(db)
	taskHandler := taskHandler.NewTaskHandler(db, reportRepo)

	mux.HandleFunc(tasks.TaskAutoResolveReport, taskHandler.AutoResolveReportHandler)
}
