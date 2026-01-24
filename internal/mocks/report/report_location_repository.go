package report

import (
	"context"
	"pingspot/internal/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockReportLocationRepository struct {
	mock.Mock
}

func (m *MockReportLocationRepository) Create(ctx context.Context, location *model.ReportLocation, tx *gorm.DB) error {
	args := m.Called(ctx, location, tx)
	return args.Error(0)
}

func (m *MockReportLocationRepository) UpdateTX(ctx context.Context, tx *gorm.DB, location *model.ReportLocation) (*model.ReportLocation, error) {
	args := m.Called(ctx, tx, location)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportLocation), args.Error(1)
}

func (m *MockReportLocationRepository) GetByReportID(ctx context.Context, reportID uint) (*model.ReportLocation, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportLocation), args.Error(1)
}
