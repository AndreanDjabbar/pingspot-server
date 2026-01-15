package report

import (
	"context"
	"pingspot/internal/domain/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockReportImageRepository struct {
	mock.Mock
}

func (m *MockReportImageRepository) Create(ctx context.Context, images *model.ReportImage, tx *gorm.DB) error {
	args := m.Called(ctx, images, tx)
	return args.Error(0)
}

func (m *MockReportImageRepository) UpdateTX(ctx context.Context, tx *gorm.DB, images *model.ReportImage) (*model.ReportImage, error) {
	args := m.Called(ctx, tx, images)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportImage), args.Error(1)
}

func (m *MockReportImageRepository) GetByReportID(ctx context.Context, reportID uint) (*model.ReportImage, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportImage), args.Error(1)
}
