package cron_worker

import (
	reportRepo "pingspot/internal/domain/report_service/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/logger"
	"pingspot/internal/task"

	"github.com/hibiken/asynq"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

func StartCron(client *asynq.Client) {
	c := cron.New(cron.WithSeconds())
	db := database.GetPostgresDB()
	reportRepo := reportRepo.NewReportRepository(db)
	tasksService := task.NewTaskService(client)

	cronHandler := newCronHandler(db, reportRepo, tasksService)

	_, err := c.AddFunc("0 0 11 * * *", func() {
		err := cronHandler.checkPotentiallyResolvedReport()
		if err != nil {
			logger.Error("Error executing CheckPotentiallyResolvedReport", zap.Error(err))
		}
	})
	if err != nil {
		logger.Error("Failed to schedule auto-resolve reports task", zap.Error(err))
	}

	_, err = c.AddFunc("0 0 12 * * *", func() {
		err := cronHandler.expireOldReports()
		if err != nil {
			logger.Error("Error executing ExpireOldReports", zap.Error(err))
		}
	})
	if err != nil {
		logger.Error("Failed to schedule expire old reports task", zap.Error(err))
	}

	_, err = c.AddFunc("0 0 12 * * *", func() {
		err := cronHandler.hardDeleteOldReports()
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