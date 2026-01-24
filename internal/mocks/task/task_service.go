package task

import (
	"github.com/stretchr/testify/mock"
)

type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) AutoResolveReportTask(reportID uint) error {
	args := m.Called(reportID)
	return args.Error(0)
}
