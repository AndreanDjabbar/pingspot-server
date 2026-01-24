package report

import (
	"context"
	"pingspot/internal/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockReportVoteRepository struct {
	mock.Mock
}

func (m *MockReportVoteRepository) GetByUserReportID(ctx context.Context, userID, reportID uint) (*model.ReportVote, error) {
	args := m.Called(ctx, userID, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportVote), args.Error(1)
}

func (m *MockReportVoteRepository) GetByUserReportIDTX(ctx context.Context, tx *gorm.DB, userID, reportID uint) (*model.ReportVote, error) {
	args := m.Called(ctx, tx, userID, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportVote), args.Error(1)
}

func (m *MockReportVoteRepository) CreateTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) (*model.ReportVote, error) {
	args := m.Called(ctx, tx, vote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportVote), args.Error(1)
}

func (m *MockReportVoteRepository) UpdateTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) (*model.ReportVote, error) {
	args := m.Called(ctx, tx, vote)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportVote), args.Error(1)
}

func (m *MockReportVoteRepository) DeleteTX(ctx context.Context, tx *gorm.DB, vote *model.ReportVote) error {
	args := m.Called(ctx, tx, vote)
	return args.Error(0)
}

func (m *MockReportVoteRepository) GetReportVoteCount(ctx context.Context, voteType model.ReportStatus, reportID uint) (int64, error) {
	args := m.Called(ctx, voteType, reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportVoteRepository) GetReportVoteCountsTX(ctx context.Context, tx *gorm.DB, reportID uint) (map[model.ReportStatus]int64, error) {
	args := m.Called(ctx, tx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[model.ReportStatus]int64), args.Error(1)
}

func (m *MockReportVoteRepository) GetHighestVoteTypeTX(ctx context.Context, tx *gorm.DB, reportID uint) (model.ReportStatus, error) {
	args := m.Called(ctx, tx, reportID)
	return args.Get(0).(model.ReportStatus), args.Error(1)
}

func (m *MockReportVoteRepository) GetResolvedVoteCount(ctx context.Context, reportID uint) (int64, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportVoteRepository) GetOnProgressVoteCount(ctx context.Context, reportID uint) (int64, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportVoteRepository) GetNotResolvedVoteCount(ctx context.Context, reportID uint) (int64, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportVoteRepository) GetTotalVoteCountTX(ctx context.Context, tx *gorm.DB, reportID uint) (int64, error) {
	args := m.Called(ctx, tx, reportID)
	return args.Get(0).(int64), args.Error(1)
}
