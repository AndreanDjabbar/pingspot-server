package task

import (
	"encoding/json"
	"fmt"
	"pingspot/internal/logger"
	"pingspot/internal/model"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskService interface {
	AutoResolveReportTask(reportID uint) error
	CreateNotificationTask(userID uint, title string, description string, entityID *string, entityType model.EntityType, category model.NotificationCategory, notificationType model.NotificationType) error
	SendFollowerReportNotificationTask(report model.Report, user model.User) error
	SendReportCommentNotificationTask(report model.Report, comment model.ReportComment, commenter model.User) error
}

type taskService struct {
	client *asynq.Client
}

func NewTaskService(client *asynq.Client) TaskService {
	return &taskService{
		client: client,
	}
}

func (s *taskService) AutoResolveReportTask(reportID uint) error {
	payload, err := json.Marshal(UpdateProgressPayload{ReportID: reportID})
	if err != nil {
		return fmt.Errorf("failed to marshal auto resolve report payload: %w", err)
	}
	task := asynq.NewTask(TaskAutoResolveReport, payload)
	_, err = s.client.Enqueue(task, asynq.ProcessIn(20*time.Minute))
	if err != nil {
		return fmt.Errorf("failed to enqueue auto resolve report task: %w", err)
	}
	logger.Info("Auto resolve report task enqueued for", zap.Int("report_id", int(reportID)))
	return nil
}

func (s *taskService) CreateNotificationTask(userID uint, title string, description string, entityID *string, entityType model.EntityType, category model.NotificationCategory, notificationType model.NotificationType) error {
	payload, err := json.Marshal(CreateNotificationPayload{
		UserID:      userID,
		Title:       title,
		Description: description,
		EntityID:    entityID,
		EntityType:  entityType,
		Category:    category,
		Type:        notificationType,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal create notification payload: %w", err)
	}
	task := asynq.NewTask(TaskCreateNotification, payload)
	_, err = s.client.Enqueue(task, asynq.ProcessIn(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to enqueue create notification task: %w", err)
	}
	logger.Info("Create notification task enqueued for", zap.Int("user_id", int(userID)))
	return nil
}

func (s *taskService) SendFollowerReportNotificationTask(report model.Report, user model.User) error {
	payload, err := json.Marshal(SendFollowerReportNotificationPayload{
		Report: report,
		User:   user,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal send follower report notification payload: %w", err)
	}
	task := asynq.NewTask(TaskSendFollowerReportNotification, payload)
	_, err = s.client.Enqueue(task, asynq.ProcessIn(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to enqueue send follower report notification task: %w", err)
	}
	logger.Info("Send follower report notification task enqueued for", zap.Int("report_id", int(report.ID)))
	return nil
}

func (s *taskService) SendReportCommentNotificationTask(report model.Report, comment model.ReportComment, commenter model.User) error {
	payload, err := json.Marshal(SendReportCommentNotificationPayload{
		Report:    report,
		Comment:   comment,
		Commenter: commenter,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal send report comment notification payload: %w", err)
	}
	task := asynq.NewTask(TaskSendReportCommentNotification, payload)
	_, err = s.client.Enqueue(task, asynq.ProcessIn(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to enqueue send report comment notification task: %w", err)
	}
	logger.Info("Send report comment notification task enqueued for", zap.Int("report_id", int(report.ID)))
	return nil
}
