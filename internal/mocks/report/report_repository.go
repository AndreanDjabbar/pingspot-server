package report

import (
	"context"
	"pingspot/internal/domain/model"
	"pingspot/internal/domain/reportService/dto"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) Create(ctx context.Context, report *model.Report, tx *gorm.DB) error {
	args := m.Called(ctx, report, tx)
	return args.Error(0)
}

func (m *MockReportRepository) UpdateTX(ctx context.Context, tx *gorm.DB, report *model.Report) (*model.Report, error) {
	args := m.Called(ctx, tx, report)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportRepository) DeleteTX(ctx context.Context, tx *gorm.DB, report *model.Report) (*model.Report, error) {
	args := m.Called(ctx, tx, report)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByID(ctx context.Context, reportID uint) (*model.Report, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByIDTX(ctx context.Context, tx *gorm.DB, reportID uint) (*model.Report, error) {
	args := m.Called(ctx, tx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportRepository) Get(ctx context.Context) (*[]model.Report, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReportStatus(ctx context.Context, status ...string) (*[]model.Report, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReportStatusCount(ctx context.Context, status ...string) (map[string]int64, error) {
	args := m.Called(ctx, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockReportRepository) GetByIDIsDeleted(ctx context.Context, reportID uint, isDeleted bool) (*model.Report, error) {
	args := m.Called(ctx, reportID, isDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByIsDeleted(ctx context.Context, isDeleted bool) ([]*model.Report, error) {
	args := m.Called(ctx, isDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByIsDeletedPaginated(ctx context.Context, limit, cursorID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance, isDeleted bool) (*[]model.Report, error) {
	args := m.Called(ctx, limit, cursorID, reportType, status, sortBy, hasProgress, distance, isDeleted)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Report), args.Error(1)
}

func (m *MockReportRepository) GetPaginated(ctx context.Context, limit, cursorID uint, reportType, status, sortBy, hasProgress string, distance dto.Distance) (*[]model.Report, error) {
	args := m.Called(ctx, limit, cursorID, reportType, status, sortBy, hasProgress, distance)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Report), args.Error(1)
}

func (m *MockReportRepository) GetByReportTypeCount(ctx context.Context) (*dto.TotalReportCount, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TotalReportCount), args.Error(1)
}

func (m *MockReportRepository) GetMonthlyReportCount(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockReportRepository) FullTextSearchReport(ctx context.Context, searchQuery string, limit int) (*[]model.Report, error) {
	args := m.Called(ctx, searchQuery, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Report), args.Error(1)
}

func (m *MockReportRepository) FullTextSearchReportPaginated(ctx context.Context, searchQuery string, limit int, cursorID uint) (*[]model.Report, error) {
	args := m.Called(ctx, searchQuery, limit, cursorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.Report), args.Error(1)
}
