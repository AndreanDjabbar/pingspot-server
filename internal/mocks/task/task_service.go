package task

import (
	"pingspot/internal/model"

	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) AutoResolveReportTask(reportID uint) error {
	args := m.Called(reportID)
	return args.Error(0)
}

func (m *MockTaskService) CreateNotificationTask(userID uint, title string, description string, entityID *string, entityType model.EntityType, category model.NotificationCategory, notificationType model.NotificationType) error {
	args := m.Called(userID, title, description, entityID, entityType, category, notificationType)
	return args.Error(0)
}
