package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"pingspot/internal/domain/reportService/repository"
	"pingspot/internal/domain/taskService/payload"
	"pingspot/internal/model"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/mainUtils"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB         *gorm.DB
	ReportRepo repository.ReportRepository
}

func NewTaskHandler(db *gorm.DB, reportRepo repository.ReportRepository) *TaskHandler {
	return &TaskHandler{
		DB:         db,
		ReportRepo: reportRepo,
	}
}

func (h *TaskHandler) AutoResolveReportHandler(ctx context.Context, t *asynq.Task) error {
	var payload payload.UpdateProgressPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}

	tx := h.DB.Begin()
	report, err := h.ReportRepo.GetByIDTX(ctx, tx, payload.ReportID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("report not found: %w", err)
	}

	if report.ReportStatus == "POTENTIALLY_RESOLVED" {
		if report.PotentiallyResolvedAt == nil {
			tx.Rollback()
			return fmt.Errorf("report %d has POTENTIALLY_RESOLVED status but PotentiallyResolvedAt is nil", report.ID)
		}

		lastUpdate := time.Unix(*report.PotentiallyResolvedAt, 0)
		if time.Since(lastUpdate) >= 20*time.Minute {
			report.ReportStatus = "RESOLVED"
			report.LastUpdatedProgressAt = mainutils.Int64PtrOrNil(time.Now().Unix())
			report.LastUpdatedBy = model.System
			if _, err := h.ReportRepo.UpdateTX(ctx, tx, report); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update report: %w", err)
			}
			tx.Commit()
			logger.Info("Auto resolve report handler success for", zap.Int("report_id", int(report.ID)))
		} else {
			tx.Rollback()
		}
	} else {
		tx.Rollback()
	}

	return nil
}
