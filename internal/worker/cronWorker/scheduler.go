package cronWorker

import (
	"pingspot/internal/domain/reportService/repository"
	tasksService "pingspot/internal/domain/taskService/service"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/worker/cronWorker/handler"
	"pingspot/pkg/logger"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func StartCron(client *asynq.Client) {
	c := cron.New(cron.WithSeconds())
	db := database.GetPostgresDB()
	reportRepo := repository.NewReportRepository(db)
	tasksService := tasksService.NewTaskService(client, reportRepo)
	
	cronHandler := handler.NewCronHandler(db, reportRepo, tasksService)

	_, err := c.AddFunc("0 0 11 * * *", func() {
		err := cronHandler.CheckPotentiallyResolvedReport()
		if err != nil {
			logger.Error("Error executing CheckPotentiallyResolvedReport", zap.Error(err))
		}
	})
	if err != nil {
		logger.Error("Failed to schedule auto-resolve reports task", zap.Error(err))
	}

	_, err = c.AddFunc("0 0 12 * * *", func() {
		err := cronHandler.ExpireOldReports()
		if err != nil {
			logger.Error("Error executing ExpireOldReports", zap.Error(err))
		}
	})
	if err != nil {
		logger.Error("Failed to schedule expire old reports task", zap.Error(err))
	}

	_, err = c.AddFunc("0 0 12 * * *", func() {
		err := cronHandler.HardDeleteOldReports()
		if err != nil {
			logger.Error("Error executing hard delete reports task", zap.Error(err))
		}
	})
	if err != nil {
		logger.Error("Failed to schedule hard delete reports task", zap.Error(err))
	}

	// _, err = c.AddFunc("0 */5 * * * *", func() {
	// })
	// if err != nil {
	// 	logger.Error("", zap.Error(err))
	// }

	c.Start()
	logger.Info("Cron scheduler started")
}