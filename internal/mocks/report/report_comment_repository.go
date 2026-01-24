package report

import (
	"context"
	"pingspot/internal/domain/model"

	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockReportCommentRepository struct {
	mock.Mock
}

func (m *MockReportCommentRepository) Create(ctx context.Context, comment *model.ReportComment) (*model.ReportComment, error) {
	args := m.Called(ctx, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportComment), args.Error(1)
}

func (m *MockReportCommentRepository) GetByID(ctx context.Context, commentID primitive.ObjectID) (*model.ReportComment, error) {
	args := m.Called(ctx, commentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ReportComment), args.Error(1)
}

func (m *MockReportCommentRepository) GetByIDs(ctx context.Context, commentIDs []primitive.ObjectID) ([]*model.ReportComment, error) {
	args := m.Called(ctx, commentIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ReportComment), args.Error(1)
}

func (m *MockReportCommentRepository) GetByReportID(ctx context.Context, reportID uint) ([]*model.ReportComment, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ReportComment), args.Error(1)
}

func (m *MockReportCommentRepository) GetCountsByReportID(ctx context.Context, reportID uint) (int64, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportCommentRepository) GetCountsByRootID(ctx context.Context, rootID primitive.ObjectID) (int64, error) {
	args := m.Called(ctx, rootID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockReportCommentRepository) GetPaginatedRootByReportID(ctx context.Context, reportID uint, cursorID *primitive.ObjectID, limit int) ([]*model.ReportComment, error) {
	args := m.Called(ctx, reportID, cursorID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ReportComment), args.Error(1)
}

func (m *MockReportCommentRepository) GetPaginatedRepliesByRootID(ctx context.Context, rootID primitive.ObjectID, cursorID *primitive.ObjectID, limit int) ([]*model.ReportComment, error) {
	args := m.Called(ctx, rootID, cursorID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*model.ReportComment), args.Error(1)
}
