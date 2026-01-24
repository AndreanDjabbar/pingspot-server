package service

import (
	"context"
	"errors"
	userMocks "pingspot/internal/mocks/user"
	"pingspot/internal/domain/model"
	"pingspot/internal/domain/userService/dto"
	mainutils "pingspot/pkg/utils/mainUtils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{}, &model.UserProfile{})
	require.NoError(t, err)

	return db
}

func TestNewUserService(t *testing.T) {
	t.Run("should create new user service", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)

		service := NewUserService(mockUserRepo, mockProfileRepo)

		assert.NotNil(t, service)
		assert.Equal(t, mockUserRepo, service.userRepo)
		assert.Equal(t, mockProfileRepo, service.userProfileRepo)
	})
}

func TestUserService_GetProfile(t *testing.T) {
	t.Run("should get user profile successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(1)
		expectedUser := &model.User{
			ID:       userID,
			FullName: "John Doe",
			Email:    "john@example.com",
			Username: "johndoe",
			Profile: model.UserProfile{
				Bio:            mainutils.StrPtrOrNil("test_bio"),
				ProfilePicture: mainutils.StrPtrOrNil("test_picture.jpg"),
				Birthday:       mainutils.StrPtrOrNil("2000-01-01"),
				Gender:         mainutils.StrPtrOrNil("male"),
			},
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		result, err := service.GetProfile(ctx, userID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.UserID)
		assert.Equal(t, "John Doe", result.FullName)
		assert.Equal(t, "john@example.com", result.Email)
		assert.Equal(t, "johndoe", result.Username)
		assert.Equal(t, mainutils.StrPtrOrNil("test_bio"), result.Bio)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(999)

		mockUserRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

		result, err := service.GetProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "tidak ditemukan")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error on database failure", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(1)
		dbError := errors.New("database connection failed")

		mockUserRepo.On("GetByID", ctx, userID).Return(nil, dbError)

		result, err := service.GetProfile(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetProfileByUsername(t *testing.T) {
	t.Run("should get user profile by username successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		username := "johndoe"
		expectedUser := &model.User{
			ID:       1,
			FullName: "John Doe",
			Email:    "john@example.com",
			Username: username,
			Profile: model.UserProfile{
				Bio:            mainutils.StrPtrOrNil("test_bio"),
				ProfilePicture: mainutils.StrPtrOrNil("test_picture.jpg"),
				Gender:         mainutils.StrPtrOrNil("male"),
			},
		}

		mockUserRepo.On("GetByUsername", ctx, username).Return(expectedUser, nil)

		result, err := service.GetProfileByUsername(ctx, username)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, username, result.Username)
		assert.Equal(t, "John Doe", result.FullName)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when username not found", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		username := "nonexistent"

		mockUserRepo.On("GetByUsername", ctx, username).Return(nil, gorm.ErrRecordNotFound)

		result, err := service.GetProfileByUsername(ctx, username)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_SaveSecurity(t *testing.T) {
	t.Run("should update password successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(1)

		currentPassword := "oldPassword123"
		hashedCurrentPassword := "$2a$10$YourHashedPasswordHere"
		user := &model.User{
			ID:       userID,
			Password: &hashedCurrentPassword,
		}

		newPassword := "newPassword456"

		req := dto.SaveUserSecurityRequest{
			CurrentPassword: currentPassword,
			NewPassword:     newPassword,
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockUserRepo.On("Save", ctx, mock.AnythingOfType("*model.User")).Return(nil)

		err := service.SaveSecurity(ctx, userID, req)

		assert.Error(t, err)
		mockUserRepo.AssertCalled(t, "GetByID", ctx, userID)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(999)
		req := dto.SaveUserSecurityRequest{
			CurrentPassword: "oldPassword",
			NewPassword:     "newPassword",
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

		err := service.SaveSecurity(ctx, userID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tidak ditemukan")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when current password is invalid", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(1)
		wrongPassword := "wrongPassword"
		hashedPassword := "$2a$10$DifferentHashedPassword"

		user := &model.User{
			ID:       userID,
			Password: &hashedPassword,
		}

		req := dto.SaveUserSecurityRequest{
			CurrentPassword: wrongPassword,
			NewPassword:     "newPassword123",
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)

		err := service.SaveSecurity(ctx, userID, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "salah")
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_SaveProfile(t *testing.T) {
	t.Run("should save user profile successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		userID := uint(1)
		req := dto.SaveUserProfileRequest{
			FullName:       "Jane Doe",
			Username:       mainutils.StrPtrOrNil("janedoe"),
			Bio:            mainutils.StrPtrOrNil("Updated bio"),
			ProfilePicture: mainutils.StrPtrOrNil("updated_picture.jpg"),
			Birthday:       mainutils.StrPtrOrNil("1995-05-15"),
			Gender:         mainutils.StrPtrOrNil("female"),
		}

		existingUser := &model.User{
			ID:       userID,
			FullName: "John Doe",
			Profile: model.UserProfile{
				Bio:            mainutils.StrPtrOrNil("Old bio"),
				ProfilePicture: mainutils.StrPtrOrNil("old_picture.jpg"),
				Birthday:       mainutils.StrPtrOrNil("1990-01-01"),
				Gender:         mainutils.StrPtrOrNil("male"),
			},
		}
		updatedProfile := &model.UserProfile{
			Bio:            req.Bio,
			ProfilePicture: req.ProfilePicture,
			Birthday:       req.Birthday,
			Gender:         req.Gender,
		}

		db := setupTestDB(t)

		mockUserRepo.On("UpdateFullNameTX", ctx, mock.Anything, userID, req.FullName).Return(nil)
		mockProfileRepo.On("GetByIDTX", ctx, mock.Anything, userID).Return(&existingUser.Profile, nil)
		mockProfileRepo.On("UpdateTX", ctx, mock.Anything, mock.AnythingOfType("*model.UserProfile")).Return(updatedProfile, nil)

		profileResponse, err := service.SaveProfile(ctx, db, userID, req)

		require.NoError(t, err)
		assert.NotNil(t, profileResponse)
		assert.Equal(t, userID, profileResponse.UserID)
		assert.Equal(t, "Jane Doe", profileResponse.FullName)
		assert.Equal(t, req.Bio, profileResponse.Bio)
		assert.Equal(t, req.ProfilePicture, profileResponse.ProfilePicture)
		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
	})
	t.Run("should create new profile when user profile not found", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)
		ctx := context.Background()
		userID := uint(1)
		req := dto.SaveUserProfileRequest{
			FullName:       "Jane Doe",
			Bio:            mainutils.StrPtrOrNil("Updated bio"),
			ProfilePicture: mainutils.StrPtrOrNil("updated_picture.jpg"),
			Birthday:       mainutils.StrPtrOrNil("1995-05-15"),
			Gender:         mainutils.StrPtrOrNil("female"),
		}

		db := setupTestDB(t)

		mockUserRepo.On("UpdateFullNameTX", ctx, mock.Anything, userID, req.FullName).Return(nil)
		mockProfileRepo.On("GetByIDTX", ctx, mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)
		mockProfileRepo.On("CreateTX", ctx, mock.Anything, mock.AnythingOfType("*model.UserProfile")).Return(nil, nil)
		
		profileResponse, err := service.SaveProfile(ctx, db, userID, req)

		require.NoError(t, err)
		assert.NotNil(t, profileResponse)
		assert.Equal(t, userID, profileResponse.UserID)
		assert.Equal(t, "Jane Doe", profileResponse.FullName)
		assert.Equal(t, req.Bio, profileResponse.Bio)
		assert.Equal(t, req.ProfilePicture, profileResponse.ProfilePicture)
		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("should return error when failed updated user profile", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)
		ctx := context.Background()
		userID := uint(1)
		req := dto.SaveUserProfileRequest{
			FullName:       "Jane Doe",
			Bio:            mainutils.StrPtrOrNil("Updated bio"),
			ProfilePicture: mainutils.StrPtrOrNil("updated_picture.jpg"),
			Birthday:       mainutils.StrPtrOrNil("1995-05-15"),
			Gender:         mainutils.StrPtrOrNil("female"),
		}

		existingUser := &model.User{
			ID:       userID,
			FullName: "John Doe",
			Profile: model.UserProfile{
				Bio:            mainutils.StrPtrOrNil("Old bio"),
				ProfilePicture: mainutils.StrPtrOrNil("old_picture.jpg"),
				Birthday:       mainutils.StrPtrOrNil("1990-01-01"),
				Gender:         mainutils.StrPtrOrNil("male"),
			},
		}

		db := setupTestDB(t)

		mockUserRepo.On("UpdateFullNameTX", ctx, mock.Anything, userID, req.FullName).Return(nil)
		mockProfileRepo.On("GetByIDTX", ctx, mock.Anything, userID).Return(&existingUser.Profile, nil)
		mockProfileRepo.On("UpdateTX", ctx, mock.Anything, mock.AnythingOfType("*model.UserProfile")).Return(nil, errors.New("update failed"))
		
		profileResponse, err := service.SaveProfile(ctx, db, userID, req)

		assert.Error(t, err)
		assert.Nil(t, profileResponse)
		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("should return error when failed create user profile", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)
		ctx := context.Background()
		userID := uint(1)
		req := dto.SaveUserProfileRequest{
			FullName:       "Jane Doe",
			Bio:            mainutils.StrPtrOrNil("Updated bio"),
			ProfilePicture: mainutils.StrPtrOrNil("updated_picture.jpg"),
			Birthday:       mainutils.StrPtrOrNil("1995-05-15"),
			Gender:         mainutils.StrPtrOrNil("female"),
		}
		db := setupTestDB(t)

		mockUserRepo.On("UpdateFullNameTX", ctx, mock.Anything, userID, req.FullName).Return(nil)
		mockProfileRepo.On("GetByIDTX", ctx, mock.Anything, userID).Return(nil, gorm.ErrRecordNotFound)
		mockProfileRepo.On("CreateTX", ctx, mock.Anything, mock.AnythingOfType("*model.UserProfile")).Return(nil, errors.New("create failed"))
		profileResponse, err := service.SaveProfile(ctx, db, userID, req)

		assert.Error(t, err)
		assert.Nil(t, profileResponse)
		assert.Contains(t, err.Error(), "gagal membuat")
		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserStatistics(t *testing.T) {
	t.Run("should get user statistics successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)

		ctx := context.Background()
		expectedStats := dto.GetUserStatisticsResponse{
			TotalUsers: 100,
			MonthlyUserCounts: map[string]int64{
				"2024-01": 20,
				"2024-02": 30,
				"2024-03": 50,
			},
			UsersByGender: map[string]int64{
				"male": 40,
				"female": 50,
				"unknown": 10,
			},
		}

		mockUserRepo.On("GetUsersCount", ctx).Return(int64(100), nil)

		mockUserRepo.On("GetByUserGenderCount", ctx).Return(map[string]int64{
			"male": 40,
			"female": 50,
			"unknown": 10,
		}, nil)

		mockUserRepo.On("GetMonthlyUserCounts", ctx).Return(map[string]int64{
			"2024-01": 20,
			"2024-02": 30,
			"2024-03": 50,
		}, nil)

		result, err := service.GetUserStatistics(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, expectedStats.TotalUsers, result.TotalUsers)
		assert.Equal(t, expectedStats.MonthlyUserCounts, result.MonthlyUserCounts)
		assert.Equal(t, expectedStats.UsersByGender, result.UsersByGender)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error on user count fetch failure", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)
		ctx := context.Background()

		mockUserRepo.On("GetUsersCount", ctx).Return(int64(0), errors.New("database error"))
		result, err := service.GetUserStatistics(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "gagal mendapatkan jumlah pengguna")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error on gender count fetch failure", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)
		ctx := context.Background()
		mockUserRepo.On("GetUsersCount", ctx).Return(int64(100), nil)
		mockUserRepo.On("GetByUserGenderCount", ctx).Return(nil, errors.New("database error"))
		result, err := service.GetUserStatistics(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "gagal mendapatkan jumlah pengguna berdasarkan gender")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error on monthly user count fetch failure", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		service := NewUserService(mockUserRepo, mockProfileRepo)
		ctx := context.Background()
		mockUserRepo.On("GetUsersCount", ctx).Return(int64(100), nil)
		mockUserRepo.On("GetByUserGenderCount", ctx).Return(map[string]int64{
			"male": 40,
			"female": 50,
			"unknown": 10,
		}, nil)
		mockUserRepo.On("GetMonthlyUserCounts", ctx).Return(nil, errors.New("database error"))
		result, err := service.GetUserStatistics(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "gagal mendapatkan jumlah pengguna bulanan")
		mockUserRepo.AssertExpectations(t)
	})
}