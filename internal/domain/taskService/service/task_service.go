package service

import (
	"encoding/json"
	"fmt"
	"pingspot/internal/domain/reportService/repository"
	"pingspot/internal/domain/taskService/payload"
	"pingspot/internal/domain/taskService/tasks"
	"pingspot/pkg/logger"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type TaskService interface {
	AutoResolveReportTask(reportID uint) error
}

type taskService struct {
	client     *asynq.Client
	ReportRepo repository.ReportRepository
}

func NewTaskService(client *asynq.Client, reportRepo repository.ReportRepository) TaskService {
	return &taskService{
		client:     client,
		ReportRepo: reportRepo,
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
