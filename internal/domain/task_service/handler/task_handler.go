package handler

import (
	"context"
	"encoding/json"
	"fmt"
	ReportRepo "pingspot/internal/domain/report_service/repository"
	NotificationRepo "pingspot/internal/domain/notification_service/repository"
	"pingspot/internal/domain/task_service/payload"
	"pingspot/internal/model"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/main_util"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB         *gorm.DB
	ReportRepo ReportRepo.ReportRepository
	NotificationRepo NotificationRepo.NotificationRepository
}

func NewTaskHandler(db *gorm.DB, reportRepo ReportRepo.ReportRepository, notificationRepo NotificationRepo.NotificationRepository) *TaskHandler {
	return &TaskHandler{
		DB:         db,
		ReportRepo: reportRepo,
		NotificationRepo: notificationRepo,
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

func (h *TaskHandler) CreateNotificationHandler(ctx context.Context, t *asynq.Task) error {
	var payload payload.CreateNotificationPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return err
	}
	Tx := h.DB.Begin()
	notification := &model.Notification{
		UserID:      payload.UserID,
		Title:       payload.Title,
		Description: payload.Description,
		EntityID:    payload.EntityID,
		EntityType:  &payload.EntityType,
		Category:    payload.Category,
		Type:        payload.Type,
		IsRead:      mainutils.BoolPtrOrNil(false),
	}

	if err := h.NotificationRepo.CreateTX(ctx, Tx, notification); err != nil {
		Tx.Rollback()
		return fmt.Errorf("failed to create notification: %w", err)
	}
	Tx.Commit()

	return nil
}