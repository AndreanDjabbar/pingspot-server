package handler

import (
	"context"
	"fmt"
	"pingspot/internal/domain/reportService/repository"
	"pingspot/internal/domain/taskService/service"
	"pingspot/internal/worker/cronWorker/util"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"
	mainutils "pingspot/pkg/utils/mainUtils"
	"time"

	"gorm.io/gorm"
)

type CronHandler struct {
	db           *gorm.DB
	reportRepo   repository.ReportRepository
	tasksService service.TaskService
}

func NewCronHandler(db *gorm.DB, reportRepo repository.ReportRepository, tasksService service.TaskService) *CronHandler {
	return &CronHandler{
		db:           db,
		reportRepo:   reportRepo,
		tasksService: tasksService,
	}
}

func (h *CronHandler) CheckPotentiallyResolvedReport() error {
	logger.Info("Executing CheckPotentiallyResolvedReport cron job")
	ctx := context.Background()
	reports, err := h.reportRepo.GetByReportStatus(ctx, "POTENTIALLY_RESOLVED")
	if err != nil {
		return fmt.Errorf("failed to get potentially resolved reports: %w", err)
	}

	for _, report := range *reports {
		if err := h.tasksService.AutoResolveReportTask(report.ID); err != nil {
			logger.Error(fmt.Sprintf("failed to enqueue auto resolve report task for report ID %d: %v", report.ID, err))
			continue
		}

		if report.User.Email == "" || report.User.Username == "" {
			logger.Error(fmt.Sprintf("user data not properly loaded for report ID %d, skipping email", report.ID))
			continue
		}
		remainingDay := util.GetAutoResolvedRemainingDay(report)
		reportLink := fmt.Sprintf("%s/main/report/%d", env.ClientURL(), report.ID)
		go util.SendAutoResolvedRemainingDayEmail(report.User.Email, report.User.Username, report.ReportTitle, reportLink, remainingDay)
	}
	return nil
}

func (h *CronHandler) ExpireOldReports() error {
	logger.Info("Executing ExpireOldReports cron job")
	ctx := context.Background()

	reports, err := h.reportRepo.GetByReportStatus(ctx, "WAITING", "ON_PROGRESS")
	if err != nil {
		return fmt.Errorf("failed to get reports for expiration check: %w", err)
	}

	now := time.Now().Unix()
	threshold := now - (30 * 24 * 60 * 60)

	for _, report := range *reports {
		if report.LastUpdatedProgressAt != nil && report.LastUpdatedBy != "" {
			if *report.LastUpdatedProgressAt <= threshold && report.LastUpdatedBy == "SYSTEM" {
				tx := h.db.Begin()

				report.ReportStatus = "EXPIRED"
				report.LastUpdatedBy = "SYSTEM"
				report.UpdatedAt = time.Now().Unix()

				if _, err := h.reportRepo.UpdateTX(ctx, tx, &report); err != nil {
					tx.Rollback()
					logger.Error(fmt.Sprintf("Failed to expire report ID %d: %v", report.ID, err))
					continue
				}

				if err := tx.Commit().Error; err != nil {
					logger.Error(fmt.Sprintf("Failed to commit transaction for report ID %d: %v", report.ID, err))
					continue
				}

				logger.Info(fmt.Sprintf("Report ID %d has been marked as EXPIRED", report.ID))
			}
		}
	}
	return nil
}

func (h *CronHandler) HardDeleteOldReports() error {
	logger.Info("Executing HardDeleteOldReports cron job")
	ctx := context.Background()

	deletedReports, err := h.reportRepo.GetByIsDeleted(ctx, true)
	if err != nil {
		return fmt.Errorf("failed to get soft-deleted reports: %w", err)
	}

	now := time.Now().Unix()
	threshold := mainutils.Int64PtrOrNil(now - (30 * 24 * 60 * 60))

	for _, deletedReport := range deletedReports {
		if deletedReport.DeletedAt != nil && threshold != nil && *deletedReport.DeletedAt <= *threshold {
			tx := h.db.Begin()

			if _, err := h.reportRepo.DeleteTX(ctx, tx, deletedReport); err != nil {
				tx.Rollback()
				logger.Error(fmt.Sprintf("Failed to hard delete report ID %d: %v", deletedReport.ID, err))
				continue
			}

			if err := tx.Commit().Error; err != nil {
				logger.Error(fmt.Sprintf("Failed to commit transaction for hard deleting report ID %d: %v", deletedReport.ID, err))
				continue
			}
			logger.Info(fmt.Sprintf("Report ID %d has been hard deleted", deletedReport.ID))
		}
	}
	return nil
}
