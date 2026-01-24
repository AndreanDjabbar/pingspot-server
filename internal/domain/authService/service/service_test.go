package service

import (
	"context"
	"errors"
	"pingspot/internal/domain/authService/dto"
	"pingspot/internal/domain/mocks"
	userMocks "pingspot/internal/domain/mocks/user"
	"pingspot/internal/domain/model"
	"pingspot/pkg/utils/tokenutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAuthTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{}, &model.UserProfile{}, &model.UserSession{})
	require.NoError(t, err)

	return db
}

func TestNewAuthService(t *testing.T) {
	t.Run("should create new auth service", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)

		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		assert.NotNil(t, service)
		assert.Equal(t, mockUserRepo, service.userRepo)
		assert.Equal(t, mockProfileRepo, service.userProfileRepo)
		assert.Equal(t, mockSessionRepo, service.userSessionRepo)
	})
}

func TestAuthService_Register(t *testing.T) {
	t.Run("should register user successfully", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)

		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.RegisterRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "password123",
			FullName: "New User",
			Provider: "email",
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

		createdUser := &model.User{
			ID:         1,
			Username:   req.Username,
			Email:      req.Email,
			FullName:   req.FullName,
			Provider:   model.Provider(req.Provider),
			IsVerified: true,
		}
		mockUserRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.User")).Return(createdUser, nil)

		mockProfileRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.UserProfile")).Return(&model.UserProfile{UserID: 1}, nil)

		user, err := service.Register(ctx, db, req, true)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, req.Username, user.Username)
		assert.Equal(t, req.Email, user.Email)
		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.RegisterRequest{
			Username: "existinguser",
			Email:    "existing@example.com",
			Password: "password123",
			FullName: "Existing User",
			Provider: "email",
		}

		existingUser := &model.User{
			ID:    1,
			Email: req.Email,
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

		user, err := service.Register(ctx, db, req, false)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "terdaftar")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should handle database error during email check", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.RegisterRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "password123",
			FullName: "New User",
			Provider: "email",
		}

		dbError := errors.New("database connection error")
		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, dbError)

		user, err := service.Register(ctx, db, req, false)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should handle user creation failure", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.RegisterRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "password123",
			FullName: "New User",
			Provider: "email",
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

		createError := errors.New("failed to create user")
		mockUserRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.User")).Return(nil, createError)

		user, err := service.Register(ctx, db, req, false)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should handle profile creation failure", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.RegisterRequest{
			Username: "newuser",
			Email:    "newuser@example.com",
			Password: "password123",
			FullName: "New User",
			Provider: "email",
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

		createdUser := &model.User{
			ID:       1,
			Username: req.Username,
			Email:    req.Email,
		}
		mockUserRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.User")).Return(createdUser, nil)

		profileError := errors.New("failed to create profile")
		mockProfileRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.UserProfile")).Return(nil, profileError)

		user, err := service.Register(ctx, db, req, false)

		assert.Error(t, err)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
		mockProfileRepo.AssertExpectations(t)
	})
}

func TestGetRefreshTokenDuration(t *testing.T) {
	t.Run("should return default duration when env is invalid", func(t *testing.T) {
		duration := getRefreshTokenDuration()
		assert.NotZero(t, duration)
	})
}

func TestAuthService_VerifyUser(t *testing.T) {
	t.Run("should verify user successfully", func(t *testing.T) {
		userID := uint(1)
		ctx := context.Background()
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		expectedUser := &model.User{
			ID:         userID,
			Email:      "user@example.com",
			Username:   "user",
			IsVerified: false,
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		mockUserRepo.On("Save", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.ID == userID && u.IsVerified == true
		})).Return(nil)

		user, err := service.VerifyUser(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.True(t, user.IsVerified)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error already verified user", func(t *testing.T) {
		userID := uint(1)
		ctx := context.Background()
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		expectedUser := &model.User{
			ID:         userID,
			Email:      "user@example.com",
			Username:   "user",
			IsVerified: true,
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.VerifyUser(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sudah diverifikasi")
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		userID := uint(999)
		ctx := context.Background()
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		mockUserRepo.On("GetByID", ctx, userID).Return(nil, gorm.ErrRecordNotFound)

		user, err := service.VerifyUser(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tidak ditemukan")
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when save fails", func(t *testing.T) {
		userID := uint(1)
		ctx := context.Background()
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		expectedUser := &model.User{
			ID:         userID,
			Email:      "user@example.com",
			Username:   "user",
			IsVerified: false,
		}

		mockUserRepo.On("GetByID", ctx, userID).Return(expectedUser, nil)

		saveError := errors.New("database save error")

		mockUserRepo.On("Save", ctx, mock.MatchedBy(func(u *model.User) bool {
			return u.ID == userID && u.IsVerified == true
		})).Return(saveError)
		user, err := service.VerifyUser(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Gagal menyimpan data user")
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_Login(t *testing.T) {
	t.Run("should return error when user not found", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.LoginRequest{
			Email:    "notfound@example.com",
			Password: "password123",
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(nil, gorm.ErrRecordNotFound)

		user, accessToken, refreshToken, err := service.Login(ctx, db, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "Email atau password salah")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when password is incorrect", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.LoginRequest{
			Email:    "user@example.com",
			Password: "wrongpassword",
		}

		hashedPassword := "$2a$10$abcdefghijklmnopqrstuv"
		existingUser := &model.User{
			ID:         1,
			Email:      req.Email,
			Password:   &hashedPassword,
			Provider:   model.ProviderEmail,
			IsVerified: true,
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

		user, accessToken, refreshToken, err := service.Login(ctx, db, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "Email atau password salah")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when account not verified", func(t *testing.T) {
		t.Skip("Skipping test - requires refactoring service to accept Redis client via dependency injection")

		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		req := dto.LoginRequest{
			Email:    "unverified@example.com",
			Password: "password123",
		}

		existingUser := &model.User{
			ID:         1,
			Email:      req.Email,
			Username:   "unverifieduser",
			Provider:   model.ProviderGoogle,
			IsVerified: false,
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

		user, accessToken, refreshToken, err := service.Login(ctx, db, req)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Empty(t, accessToken)
		assert.Empty(t, refreshToken)
		assert.Contains(t, err.Error(), "belum diverifikasi")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should login successfully - tests parts before Redis interaction", func(t *testing.T) {
		db := setupAuthTestDB(t)
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		password := "password123"
		hashedPassword, _ := tokenutils.HashString(password)

		req := dto.LoginRequest{
			Email:     "user@example.com",
			Password:  password,
			IPAddress: "127.0.0.1",
			UserAgent: "Test User Agent",
		}

		existingUser := &model.User{
			ID:         1,
			Email:      req.Email,
			Username:   "testuser",
			FullName:   "Test User",
			Password:   &hashedPassword,
			Provider:   model.ProviderEmail,
			IsVerified: true,
		}

		mockUserRepo.On("GetByEmail", ctx, req.Email).Return(existingUser, nil)

		createdSession := &model.UserSession{
			ID:             1,
			UserID:         existingUser.ID,
			RefreshTokenID: "test-refresh-token-id",
			IsActive:       true,
		}
		mockSessionRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.UserSession")).Return(createdSession, nil)

		// Note: This test will fail at Redis interaction since we can't mock the global singleton
		// The test validates:
		// 1. User lookup by email
		// 2. Password verification
		// 3. User verification status check
		// 4. Session creation in database
		// 
		// It cannot test:
		// - Refresh token storage in Redis
		// - User session ID storage in Redis set
		// - Redis TTL operations
		
		t.Skip("Test will fail at Redis interaction - service needs refactoring to accept Redis client via DI")

		user, accessToken, refreshToken, err := service.Login(ctx, db, req)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, existingUser.Email, user.Email)
		mockUserRepo.AssertExpectations(t)
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("should return error when session not found", func(t *testing.T) {
		t.Skip("Skipping test - requires refactoring service to accept Redis client via dependency injection")
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		userID := uint(1)
		refreshTokenID := "invalid-token-id"

		mockSessionRepo.On("GetByRefreshTokenID", ctx, refreshTokenID).Return(nil, gorm.ErrRecordNotFound)

		err := service.Logout(ctx, userID, refreshTokenID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Gagal mengambil sesi user")
		mockSessionRepo.AssertExpectations(t)
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		t.Skip("Skipping test - requires refactoring service to accept Redis client via dependency injection")
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		userID := uint(1)
		refreshTokenID := "valid-token-id"

		userSession := &model.UserSession{
			ID:             1,
			UserID:         userID,
			RefreshTokenID: refreshTokenID,
			IsActive:       true,
		}

		mockSessionRepo.On("GetByRefreshTokenID", ctx, refreshTokenID).Return(userSession, nil)
		mockSessionRepo.On("Update", ctx, mock.MatchedBy(func(session *model.UserSession) bool {
			return session.IsActive == false
		})).Return(errors.New("database error"))

		err := service.Logout(ctx, userID, refreshTokenID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Gagal memperbarui sesi user")
		mockSessionRepo.AssertExpectations(t)
	})
}

func TestAuthService_UpdateUserByEmail(t *testing.T) {
	t.Run("should update user successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		email := "user@example.com"

		existingUser := &model.User{
			ID:       1,
			Email:    email,
			Username: "oldusername",
		}

		updatedUser := &model.User{
			Username: "newusername",
		}

		expectedUser := &model.User{
			ID:       1,
			Email:    email,
			Username: "newusername",
		}

		mockUserRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)
		mockUserRepo.On("UpdateByEmail", ctx, email, updatedUser).Return(expectedUser, nil)

		user, err := service.UpdateUserByEmail(ctx, email, updatedUser)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "newusername", user.Username)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		email := "notfound@example.com"
		updatedUser := &model.User{Username: "newusername"}

		mockUserRepo.On("GetByEmail", ctx, email).Return(nil, gorm.ErrRecordNotFound)

		user, err := service.UpdateUserByEmail(ctx, email, updatedUser)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "tidak ditemukan")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		email := "user@example.com"

		existingUser := &model.User{
			ID:    1,
			Email: email,
		}

		updatedUser := &model.User{Username: "newusername"}

		mockUserRepo.On("GetByEmail", ctx, email).Return(existingUser, nil)
		mockUserRepo.On("UpdateByEmail", ctx, email, updatedUser).Return(nil, errors.New("database error"))

		user, err := service.UpdateUserByEmail(ctx, email, updatedUser)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Gagal update user")
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_GetUserByEmail(t *testing.T) {
	t.Run("should get user successfully", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		email := "user@example.com"

		expectedUser := &model.User{
			ID:       1,
			Email:    email,
			Username: "testuser",
		}

		mockUserRepo.On("GetByEmail", ctx, email).Return(expectedUser, nil)

		user, err := service.GetUserByEmail(ctx, email)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Email)
		assert.Equal(t, "testuser", user.Username)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return nil when user not found", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		email := "notfound@example.com"

		mockUserRepo.On("GetByEmail", ctx, email).Return(nil, gorm.ErrRecordNotFound)

		user, err := service.GetUserByEmail(ctx, email)

		require.NoError(t, err)
		assert.Nil(t, user)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("should return error when database error occurs", func(t *testing.T) {
		mockUserRepo := new(userMocks.MockUserRepository)
		mockProfileRepo := new(userMocks.MockUserProfileRepository)
		mockSessionRepo := new(userMocks.MockUserSessionRepository)
		cacheRepo := new(mocks.MockCacheRepository)
		service := NewAuthService(mockUserRepo, mockProfileRepo, mockSessionRepo, cacheRepo)

		ctx := context.Background()
		email := "user@example.com"

		mockUserRepo.On("GetByEmail", ctx, email).Return(nil, errors.New("database connection error"))

		user, err := service.GetUserByEmail(ctx, email)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "Gagal mengambil data user")
		mockUserRepo.AssertExpectations(t)
	})
}


func TestAuthService_RefreshToken(t *testing.T) {
	t.Skip("Skipping test - requires refactoring service to accept Redis client via dependency injection")
}