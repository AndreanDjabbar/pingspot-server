package report

import (
	"context"
	"pingspot/internal/domain/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockReportReactionRepository struct {
	mock.Mock
}

func (m *MockReportReactionRepository) GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportReaction, error) {
	args := m.Called(ctx, userID, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportReaction), args.Error(1)
}

func (m *MockReportReactionRepository) GetByUserReportIDTX(ctx context.Context, tx *gorm.DB, userID, reportID uint) (*model.ReportReaction, error) {
	args := m.Called(ctx, tx, userID, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportReaction), args.Error(1)
}

func (m *MockReportReactionRepository) CreateTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) (*model.ReportReaction, error) {
	args := m.Called(ctx, tx, reaction)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportReaction), args.Error(1)
}

func (m *MockReportReactionRepository) UpdateTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) (*model.ReportReaction, error) {
	args := m.Called(ctx, tx, reaction)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportReaction), args.Error(1)
}

func (m *MockReportReactionRepository) DeleteTX(ctx context.Context, tx *gorm.DB, reaction *model.ReportReaction) error {
	args := m.Called(ctx, tx, reaction)
	return args.Error(0)
}

func (m *MockReportReactionRepository) GetLikeReactionCount(ctx context.Context, reportID uint) (int64, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportReactionRepository) GetDislikeReactionCount(ctx context.Context, reportID uint) (int64, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(int64), args.Error(1)
}
