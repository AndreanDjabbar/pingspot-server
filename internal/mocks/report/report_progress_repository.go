package report

import (
	"context"
	"pingspot/internal/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockReportProgressRepository struct {
	mock.Mock
}

func (m *MockReportProgressRepository) Create(ctx context.Context, progress *model.ReportProgress) (*model.ReportProgress, error) {
	args := m.Called(ctx, progress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportProgress), args.Error(1)
}

func (m *MockReportProgressRepository) CreateTX(ctx context.Context, tx *gorm.DB, progress *model.ReportProgress) (*model.ReportProgress, error) {
	args := m.Called(ctx, tx, progress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportProgress), args.Error(1)
}

func (m *MockReportProgressRepository) GetByReportID(ctx context.Context, reportID uint) ([]model.ReportProgress, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.ReportProgress), args.Error(1)
}
