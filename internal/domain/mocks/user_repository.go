package mocks

import (
	"context"
	"pingspot/internal/domain/model"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) UpdateByEmail(ctx context.Context, email string, updatedUser *model.User) (*model.User, error) {
	args := m.Called(ctx, email, updatedUser)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID uint) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByIDs(ctx context.Context, userIDs []uint) ([]model.User, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) GetByUserGenderCount(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockUserRepository) GetUsersCount(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetMonthlyUserCounts(ctx context.Context) (map[string]int64, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]int64), args.Error(1)
}

func (m *MockUserRepository) Save(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CreateTX(ctx context.Context, tx *gorm.DB, user *model.User) (*model.User, error) {
	args := m.Called(ctx, tx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FullTextSearchUsers(ctx context.Context, query string, limit int) (*[]model.User, error) {
	args := m.Called(ctx, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (m *MockUserRepository) FullTextSearchUsersPaginated(ctx context.Context, searchQuery string, limit int, cursorID uint) (*[]model.User, error) {
	args := m.Called(ctx, searchQuery, limit, cursorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.User), args.Error(1)
}

func (m *MockUserRepository) UpdateFullNameTX(ctx context.Context, tx *gorm.DB, userID uint, fullName string) error {
	args := m.Called(ctx, tx, userID, fullName)
	return args.Error(0)
}

func (m *MockUserRepository) Get(ctx context.Context) (*[]model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*[]model.User), args.Error(1)
}