package service

import (
	"encoding/json"
	"fmt"
	"pingspot/internal/domain/task_service/payload"
	"pingspot/internal/domain/task_service/tasks"
	"pingspot/internal/model"
	"pingspot/pkg/logger"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskService interface {
	AutoResolveReportTask(reportID uint) error
	CreateNotificationTask(userID uint, title string, description string, entityID *string, entityType model.EntityType, category model.NotificationCategory, notificationType model.NotificationType) error
}

type taskService struct {
	client     *asynq.Client
}

func NewTaskService(client *asynq.Client) TaskService {
	return &taskService{
		client:     client,
	}
}

func (s *taskService) AutoResolveReportTask(reportID uint) error {
	payload, _ := json.Marshal(payload.UpdateProgressPayload{ReportID: reportID})
	task := asynq.NewTask(tasks.TaskAutoResolveReport, payload)
	_, err := s.client.Enqueue(task, asynq.ProcessIn(20*time.Minute))
	if err != nil {
		return fmt.Errorf("failed to enqueue auto resolve report task: %w", err)
	}
	logger.Info("Auto resolve report task enqueued for", zap.Int("report_id", int(reportID)))
	return nil
}

func (s *taskService) CreateNotificationTask(userID uint, title string, description string, entityID *string, entityType model.EntityType, category model.NotificationCategory, notificationType model.NotificationType) error {
	payload, _ := json.Marshal(payload.CreateNotificationPayload{
		UserID:      userID,
		Title:       title,
		Description: description,
		EntityID:    entityID,
		EntityType:  entityType,
		Category:    category,
		Type:          notificationType,
	})
	task := asynq.NewTask(tasks.TaskCreateNotification, payload)
	_, err := s.client.Enqueue(task, asynq.ProcessIn(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to enqueue create notification task: %w", err)
	}
	logger.Info("Create notification task enqueued for", zap.Int("user_id", int(userID)))
	return nil
}