package handler

import (
	"context"
	"encoding/json"
	"fmt"
	notificationRepo "pingspot/internal/domain/notification_service/repository"
	reportRepo "pingspot/internal/domain/report_service/repository"
	socialRepo "pingspot/internal/domain/social_service/repository"
	"pingspot/internal/domain/task_service/payload"
	"pingspot/internal/domain/task_service/util"
	"pingspot/internal/model"
	cacheRepo "pingspot/internal/repository"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env_util"
	mainutils "pingspot/pkg/utils/main_util"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TaskHandler struct {
	DB         *gorm.DB
	ReportRepo reportRepo.ReportRepository
	NotificationRepo notificationRepo.NotificationRepository
	SocialRepo socialRepo.FollowRepository
	CacheRepo cacheRepo.CacheRepository
}

func NewTaskHandler(db *gorm.DB, cacheRepo cacheRepo.CacheRepository, reportRepo reportRepo.ReportRepository, notificationRepo notificationRepo.NotificationRepository, socialRepo socialRepo.FollowRepository) *TaskHandler {
	return &TaskHandler{
		DB:         db,
		ReportRepo: reportRepo,
		NotificationRepo: notificationRepo,
		SocialRepo: socialRepo,
		CacheRepo: cacheRepo,
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

func (h *TaskHandler) SendFollowerReportNotificationHandler(ctx context.Context, t *asynq.Task) error {
    var payload payload.SendFollowerReportNotificationPayload
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }

    followers, err := h.SocialRepo.GetFollowersByUserID(ctx, payload.User.ID)
    if err != nil {
        return fmt.Errorf("failed to get followers: %w", err)
    }

    if len(followers) == 0 {
        logger.Info("No followers found for user", zap.Int("user_id", int(payload.User.ID)))
        return nil
    }

    const cooldownDuration = 10 * time.Hour

    var activeFollowerIDs []string
    followerMap := make(map[string]model.User)

    for _, follower := range followers {
        if follower.IsDisableEmailNotification {
            continue
        }

        idStr := fmt.Sprintf("%d", follower.ID)
        activeFollowerIDs = append(activeFollowerIDs, idStr)
        followerMap[idStr] = *follower
    }

    if len(activeFollowerIDs) == 0 {
        logger.Info("All followers have disabled email notifications", zap.Int("user_id", int(payload.User.ID)))
        return nil
    }

    allowedFollowerIDs, err := h.CacheRepo.FilterByCooldown(
        ctx, 
        "email:follower_new_report", 
        activeFollowerIDs, 
        fmt.Sprintf("%d", payload.User.ID), 
        cooldownDuration,
    )
    if err != nil {
        return fmt.Errorf("failed to filter followers by cooldown: %w", err)
    }

    if len(allowedFollowerIDs) == 0 {
        logger.Info("No followers allowed to receive notification due to cooldown", zap.Int("user_id", int(payload.User.ID)))
        return nil
    }

    eligibleFollowers := make([]model.User, len(allowedFollowerIDs))
    for i, idStr := range allowedFollowerIDs {
        eligibleFollowers[i] = followerMap[idStr]
    }

    reportLink := fmt.Sprintf("%s/main/reports/%d", env_util.ClientURL(), payload.Report.ID)

    if err := util.SendNotificationNewReportEmails(payload.Report, payload.User, eligibleFollowers, reportLink); err != nil {
        logger.Error("Failed to send report notification emails", zap.Error(err))
        return fmt.Errorf("failed to send notification emails: %w", err)
    }

    return nil
}

func (h *TaskHandler) SendReportCommentNotificationHandler(ctx context.Context, t *asynq.Task) error {
	var payload payload.SendReportCommentNotificationPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if payload.Report.User.IsDisableEmailNotification {
		logger.Info("Report owner has disabled email notifications", zap.Int("user_id", int(payload.Report.UserID)), zap.Int("report_id", int(payload.Report.ID)))
		return nil
	}

	const cooldownDuration = 4 * time.Hour

	allowedIDs, err := h.CacheRepo.FilterByCooldown(
        ctx,
        "email:report_comment",
        []string{fmt.Sprintf("%d", payload.Report.UserID)},
        fmt.Sprintf("%d", payload.Report.ID),
        cooldownDuration,
    )
    if err != nil {
        return fmt.Errorf("failed to filter comment cooldown: %w", err)
    }

    if len(allowedIDs) == 0 {
        logger.Info("Comment email suppressed due to cooldown", zap.Int("user_id", int(payload.Report.UserID)), zap.Int("report_id", int(payload.Report.ID)))
		return nil
	}

	reportLink := fmt.Sprintf("%s/main/reports/%d", env_util.ClientURL(), payload.Report.ID)

	if err := util.SendNotificationReportCommentEmail(payload.Report, payload.Comment, payload.Commenter, reportLink); err != nil {
		logger.Error("Failed to send report comment notification emails", zap.Error(err))
		return fmt.Errorf("failed to send report comment notification emails: %w", err)
	}

	return nil
}