package user

import (
	"context"
	"pingspot/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockUserSessionRepository struct {
	mock.Mock
}

func (m *MockUserSessionRepository) Update(ctx context.Context, session *model.UserSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *MockUserSessionRepository) CreateTX(ctx context.Context, tx *gorm.DB, session *model.UserSession) (*model.UserSession, error) {
	args := m.Called(ctx, tx, session)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserSession), args.Error(1)
}

func (m *MockUserSessionRepository) GetByRefreshTokenID(ctx context.Context, refreshTokenID string) (*model.UserSession, error) {
	args := m.Called(ctx, refreshTokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserSession), args.Error(1)
}

func (m *MockUserSessionRepository) DeleteTX(ctx context.Context, tx *gorm.DB, sessionID uint) error {
	args := m.Called(ctx, tx, sessionID)
	return args.Error(0)
}

func (m *MockUserSessionRepository) GetByID(ctx context.Context, sessionID uint) (*model.UserSession, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserSession), args.Error(1)
}

func (m *MockUserSessionRepository) GetByUserID(ctx context.Context, userID uint) (*[]model.UserSession, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.UserSession), args.Error(1)
}

func (m *MockUserSessionRepository) DeleteByUserIDTX(ctx context.Context, tx *gorm.DB, userID uint) error {
	args := m.Called(ctx, tx, userID)
	return args.Error(0)
}