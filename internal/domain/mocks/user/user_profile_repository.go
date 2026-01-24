package user

import (
	"context"
	"pingspot/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockUserProfileRepository struct {
	mock.Mock
}

func (m *MockUserProfileRepository) GetByIDTX(ctx context.Context, tx *gorm.DB, userID uint) (*model.UserProfile, error) {
	args := m.Called(ctx, tx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) GetByID(ctx context.Context, userID uint) (*model.UserProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) SaveByID(ctx context.Context, userID uint, profile *model.UserProfile) (*model.UserProfile, error) {
	args := m.Called(ctx, userID, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) CreateTX(ctx context.Context, tx *gorm.DB, profile *model.UserProfile) (*model.UserProfile, error) {
	args := m.Called(ctx, tx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserProfile), args.Error(1)
}

func (m *MockUserProfileRepository) UpdateTX(ctx context.Context, tx *gorm.DB, profile *model.UserProfile) (*model.UserProfile, error) {
	args := m.Called(ctx, tx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserProfile), args.Error(1)
}