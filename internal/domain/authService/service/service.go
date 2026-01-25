package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"pingspot/internal/domain/authService/dto"
	"pingspot/internal/domain/authService/util"
	userRepo "pingspot/internal/domain/userService/repository"
	"pingspot/internal/model"
	cacheRepo "pingspot/internal/repository"
	apperror "pingspot/pkg/apperror"
	"pingspot/pkg/logger"
	contextutils "pingspot/pkg/utils/contextUtils"
	"pingspot/pkg/utils/env"
	tokenutils "pingspot/pkg/utils/tokenutils"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func getRefreshTokenDuration() time.Duration {
	ageStr := env.RefreshTokenAge()
	age, err := strconv.Atoi(ageStr)
	if err != nil || age <= 0 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(age) * time.Second
}

type AuthService struct {
	userRepo        userRepo.UserRepository
	userSessionRepo userRepo.UserSessionRepository
	userProfileRepo userRepo.UserProfileRepository
	cacheRepo       cacheRepo.CacheRepository
}

func NewAuthService(
	userRepo userRepo.UserRepository,
	userProfileRepo userRepo.UserProfileRepository,
	userSessionRepo userRepo.UserSessionRepository,
	cacheRepo cacheRepo.CacheRepository,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		userSessionRepo: userSessionRepo,
		cacheRepo:       cacheRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, db *gorm.DB, req dto.RegisterRequest, isVerified bool) (*model.User, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("Registering new user",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
		zap.String("provider", req.Provider),
	)

	tx := db.Begin()
	if tx.Error != nil {
		logger.Error("Failed to start transaction",
			zap.String("request_id", requestID),
			zap.Error(tx.Error),
		)
		return nil, apperror.New(500, "TRANSACTION_START_FAILED", "Gagal memulai transaksi", tx.Error.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	_, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil {
		tx.Rollback()
		return nil, apperror.New(400, "EMAIL_ALREADY_REGISTERED", "Email sudah terdaftar", "")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, apperror.New(500, "USER_CHECK_FAILED", "Terjadi kesalahan saat cek data user", err.Error())
	}

	hashedPassword, err := tokenutils.HashString(req.Password)
	if err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "PASSWORD_HASH_FAILED", "Gagal mengenkripsi password", err.Error())
	}

	user := model.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   &hashedPassword,
		FullName:   req.FullName,
		Provider:   model.Provider(req.Provider),
		ProviderID: req.ProviderID,
		IsVerified: isVerified,
	}

	createdUser, err := s.userRepo.CreateTX(ctx, tx, &user)
	if err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "USER_CREATE_FAILED", "Gagal membuat user", err.Error())
	}

	newProfile := model.UserProfile{
		UserID: createdUser.ID,
	}

	if _, err := s.userProfileRepo.CreateTX(ctx, tx, &newProfile); err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "PROFILE_CREATE_FAILED", "Gagal membuat profil user", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit transaction",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, apperror.New(500, "TRANSACTION_COMMIT_FAILED", "Gagal menyimpan perubahan", err.Error())
	}

	logger.Info("User registered successfully",
		zap.String("request_id", requestID),
		zap.Uint("user_id", createdUser.ID),
		zap.String("email", createdUser.Email),
	)

	return createdUser, nil
}

func (s *AuthService) Login(ctx context.Context, db *gorm.DB, req dto.LoginRequest) (*model.User, string, string, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("User login attempt",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
	)

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", apperror.New(401, "INVALID_CREDENTIALS", "Email atau password salah", err.Error())
	}

	if user.Provider == model.ProviderEmail {
		if user.Password == nil || !tokenutils.CheckHashString(req.Password, *user.Password) {
			return nil, "", "", apperror.New(401, "INVALID_CREDENTIALS", "Email atau password salah", "")
		}
	}

	if !user.IsVerified {
		randomCode1, err := tokenutils.GenerateRandomCode(150)
		if err != nil {
			logger.Error("Failed to generate random code",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode acak", err.Error())
		}
		randomCode2, err := tokenutils.GenerateRandomCode(150)
		if err != nil {
			logger.Error("Failed to generate random code",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode acak", err.Error())
		}
		verificationLink := fmt.Sprintf("%s/auth/verify-account/%s/%d/%s", env.ClientURL(), randomCode1, user.ID, randomCode2)

		linkData := map[string]string{
			"link1": randomCode1,
			"link2": randomCode2,
		}
		linkJSON, err := json.Marshal(linkData)
		if err != nil {
			logger.Error("Failed to marshal verification link data",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", apperror.New(500, "VERIFICATION_CODE_SAVE_FAILED", "Gagal menyimpan kode verifikasi", err.Error())
		}
		redisKey := fmt.Sprintf("link:%d", user.ID)
		err = s.cacheRepo.Set(context.Background(), redisKey, linkJSON, 5*time.Minute)
		if err != nil {
			logger.Error("Failed to save verification link to Redis",
				zap.String("request_id", requestID),
				zap.Error(err),
			)
			return nil, "", "", apperror.New(500, "VERIFICATION_CODE_REDIS_FAILED", "Gagal menyimpan kode verifikasi ke Redis", err.Error())
		}
		go util.SendVerificationEmail(user.Email, user.Username, verificationLink)
		return nil, "", "", apperror.New(403, "ACCOUNT_NOT_VERIFIED", "Akun belum diverifikasi, silakan cek email untuk verifikasi", "")
	}

	refreshTokenID := uuid.New().String()
	refreshToken := tokenutils.GenerateRefreshToken(user.ID, refreshTokenID)
	hashedRefreshToken := tokenutils.HashSHA256String(refreshToken)

	tx := db.Begin()
	userSession, err := s.userSessionRepo.CreateTX(ctx, tx, &model.UserSession{
		UserID:             user.ID,
		ExpiresAt:          time.Now().Add(getRefreshTokenDuration()).Unix(),
		IsActive:           true,
		RefreshTokenID:     refreshTokenID,
		HashedRefreshToken: hashedRefreshToken,
		IPAddress:          req.IPAddress,
		UserAgent:          req.UserAgent,
	})
	if err != nil {
		tx.Rollback()
		logger.Error("Failed to create user session",
			zap.String("request_id", requestID),
			zap.Uint("user_id", user.ID),
			zap.Error(err),
		)
		return nil, "", "", apperror.New(500, "USER_SESSION_CREATE_FAILED", "Gagal membuat sesi user", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		logger.Error("Failed to commit user session transaction",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", apperror.New(500, "USER_SESSION_COMMIT_FAILED", "Gagal menyimpan sesi user", err.Error())
	}

	accessToken := tokenutils.GenerateAccessToken(user.ID, userSession.ID, user.Email, user.Username, user.FullName)

	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
	err = s.cacheRepo.Set(context.Background(), refreshKey, hashedRefreshToken, getRefreshTokenDuration())
	if err != nil {
		logger.Error("Failed to save refresh token to Redis",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", apperror.New(500, "REFRESH_TOKEN_SAVE_FAILED", "Gagal menyimpan refresh token", err.Error())
	}

	userSessionKey := fmt.Sprintf("user_session:%d", user.ID)
	err = s.cacheRepo.SAdd(context.Background(), userSessionKey, userSession.ID)
	if err != nil {
		logger.Error("Failed to save user session ID to Redis set",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", apperror.New(500, "USER_SESSION_SAVE_FAILED", "Gagal menyimpan sesi user", err.Error())
	}

	_, err = s.cacheRepo.Expire(context.Background(), userSessionKey, getRefreshTokenDuration())
	if err != nil {
		logger.Warn("Failed to set TTL on user session set",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
	}

	logger.Info("User logged in successfully",
		zap.String("request_id", requestID),
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
	)

	return user, accessToken, refreshToken, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := tokenutils.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", apperror.New(401, "INVALID_REFRESH_TOKEN", "Refresh token tidak valid", err.Error())
	}
	
	userID := uint(claims["user_id"].(float64))
	refreshTokenID := claims["refresh_token_id"].(string)
	hashedRefreshToken := tokenutils.HashSHA256String(refreshToken)

	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)

	storedHashedRefreshToken, err := s.cacheRepo.Get(context.Background(), refreshKey)

	var userSession *model.UserSession

	if err != nil {
		logger.Warn("Refresh token not found in Redis, checking PostgreSQL",
			zap.String("refresh_token_id", refreshTokenID),
			zap.Error(err),
		)

		userSession, err = s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
		if err != nil {
			return "", "", apperror.New(401, "USER_SESSION_NOT_FOUND", "Sesi user tidak ditemukan", err.Error())
		}

		if !userSession.IsActive || time.Now().Unix() > userSession.ExpiresAt {
			return "", "", apperror.New(401, "SESSION_INVALID", "Sesi sudah tidak aktif atau kedaluwarsa", "")
		}

		if userSession.HashedRefreshToken != hashedRefreshToken {
			return "", "", apperror.New(401, "REFRESH_TOKEN_INVALID", "Refresh token tidak cocok", "")
		}

		remainingDuration := time.Until(time.Unix(userSession.ExpiresAt, 0))
		if remainingDuration > 0 {
			err = s.cacheRepo.Set(context.Background(), refreshKey, hashedRefreshToken, remainingDuration)
			if err != nil {
				logger.Warn("Failed to restore refresh token to Redis", zap.Error(err))
			}
		}

		storedHashedRefreshToken = userSession.HashedRefreshToken
	} else if storedHashedRefreshToken != hashedRefreshToken {
		return "", "", apperror.New(401, "REFRESH_TOKEN_INVALID", "Refresh token tidak cocok", "")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return "", "", apperror.New(500, "USER_FETCH_FAILED", "Gagal mengambil data user", err.Error())
	}

	if userSession == nil {
		userSession, err = s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
		if err != nil {
			return "", "", apperror.New(401, "USER_SESSION_NOT_FOUND", "Sesi user tidak ditemukan", err.Error())
		}
	}

	if !userSession.IsActive || time.Now().Unix() > userSession.ExpiresAt {
		return "", "", apperror.New(401, "SESSION_INVALID", "Sesi sudah tidak aktif atau kedaluwarsa", "")
	}

	newRefreshTokenID := uuid.New().String()
	newRefreshToken := tokenutils.GenerateRefreshToken(userID, newRefreshTokenID)
	newHashedRefreshToken := tokenutils.HashSHA256String(newRefreshToken)

	if err := s.cacheRepo.Del(context.Background(), refreshKey); err != nil {
		logger.Warn("Failed to delete old refresh token", zap.Error(err))
	}

	newRefreshKey := fmt.Sprintf("refresh_token:%s", newRefreshTokenID)
	if err := s.cacheRepo.Set(context.Background(), newRefreshKey, newHashedRefreshToken, getRefreshTokenDuration()); err != nil {
		return "", "", apperror.New(500, "REFRESH_TOKEN_SAVE_FAILED", "Gagal menyimpan refresh token baru", err.Error())
	}

	userSession.RefreshTokenID = newRefreshTokenID
	userSession.HashedRefreshToken = newHashedRefreshToken
	userSession.ExpiresAt = time.Now().Add(getRefreshTokenDuration()).Unix()

	if err := s.userSessionRepo.Update(ctx, userSession); err != nil {
		return "", "", apperror.New(500, "SESSION_UPDATE_FAILED", "Gagal memperbarui sesi", err.Error())
	}

	accessToken := tokenutils.GenerateAccessToken(userID, userSession.ID, user.Email, user.Username, user.FullName)

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, userID uint, refreshTokenID string) error {
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
	if err := s.cacheRepo.Del(context.Background(), refreshKey); err != nil {
		logger.Warn("Failed to delete refresh token from Redis", zap.Error(err))
	}
	userSession, err := s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
	if err != nil {
		return apperror.New(500, "USER_SESSION_FETCH_FAILED", "Gagal mengambil sesi user", err.Error())
	}
	userSession.IsActive = false
	if err := s.userSessionRepo.Update(ctx, userSession); err != nil {
		return apperror.New(500, "USER_SESSION_UPDATE_FAILED", "Gagal memperbarui sesi user", err.Error())
	}
	userSessionKey := fmt.Sprintf("user_session:%d", userID)
	if err := s.cacheRepo.SRem(context.Background(), userSessionKey, userSession.ID); err != nil {
		logger.Warn("Failed to remove user session ID from Redis set", zap.Error(err))
	}
	return nil
}

func (s *AuthService) VerifyUser(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "User tidak ditemukan", err.Error())
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "Gagal mengambil data user", err.Error())
	}

	if user.IsVerified {
		return nil, apperror.New(400, "ALREADY_VERIFIED", "Akun sudah diverifikasi", "")
	}

	user.IsVerified = true

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, apperror.New(500, "USER_SAVE_FAILED", "Gagal menyimpan data user", err.Error())
	}

	return user, nil
}

func (s *AuthService) UpdateUserByEmail(ctx context.Context, email string, updatedUser *model.User) (*model.User, error) {
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "User tidak ditemukan", err.Error())
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "Gagal mencari user", err.Error())
	}

	user, err := s.userRepo.UpdateByEmail(ctx, email, updatedUser)
	if err != nil {
		return nil, apperror.New(500, "USER_UPDATE_FAILED", "Gagal update user", err.Error())
	}

	return user, nil
}

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "Gagal mengambil data user", err.Error())
	}
	return user, nil
}
