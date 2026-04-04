package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"pingspot/internal/domain/auth_service/dto"
	"pingspot/internal/domain/auth_service/util"
	userRepo "pingspot/internal/domain/user_service/repository"
	"pingspot/internal/model"
	cacheRepo "pingspot/internal/repository"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"
	contextutils "pingspot/pkg/utils/context_util"
	env "pingspot/pkg/utils/env_util"
	tokenutils "pingspot/pkg/utils/token_util"
	"strconv"
	"strings"
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
	db              *gorm.DB
	userRepo        userRepo.UserRepository
	userSessionRepo userRepo.UserSessionRepository
	userProfileRepo userRepo.UserProfileRepository
	cacheRepo       cacheRepo.CacheRepository
}

func NewAuthService(
	db *gorm.DB,
	userRepo userRepo.UserRepository,
	userProfileRepo userRepo.UserProfileRepository,
	userSessionRepo userRepo.UserSessionRepository,
	cacheRepo cacheRepo.CacheRepository,
) *AuthService {
	return &AuthService{
		db:              db,
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		userSessionRepo: userSessionRepo,
		cacheRepo:       cacheRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest, isVerified bool) (*model.User, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("Registering new user",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
		zap.String("provider", req.Provider),
	)

	tx := s.db.Begin()
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

func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*model.User, string, string, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("User login attempt",
		zap.String("request_id", requestID),
		zap.String("email", req.Email),
	)

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", "", apperror.New(401, "INVALID_CREDENTIALS", "Email atau password salah", err.Error())
	}

	if model.Provider(req.Provider) == model.ProviderEmail {
		if !tokenutils.CheckHashString(req.Password, *user.Password) {
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

	tx := s.db.Begin()
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

	sessionKey := fmt.Sprintf("session:%d", userSession.ID)
	err = s.cacheRepo.Set(context.Background(), sessionKey, user.ID, getRefreshTokenDuration())
	if err != nil {
		logger.Error("Failed to save session data to Redis",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, "", "", apperror.New(500, "SESSION_SAVE_FAILED", "Gagal menyimpan data sesi", err.Error())
	}

	accessToken := tokenutils.GenerateAccessToken(user.ID, userSession.ID, user.Email, user.Username, user.FullName)

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

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return "", "", apperror.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Claim user_id tidak valid", "")
	}

	refreshTokenID, ok := claims["refresh_token_id"].(string)
	if !ok || refreshTokenID == "" {
		return "", "", apperror.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Claim refresh_token_id tidak valid", "")
	}

	userID := uint(userIDFloat)
	hashedRefreshToken := tokenutils.HashSHA256String(refreshToken)
	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)

	storedHashedRefreshToken, err := s.cacheRepo.Get(ctx, refreshKey)
	if err != nil {
		logger.Warn("Failed to get refresh token from Redis, fallback to PostgreSQL",
			zap.String("refresh_token_id", refreshTokenID),
			zap.Error(err),
		)
	} else {
		if storedHashedRefreshToken != hashedRefreshToken {
			return "", "", apperror.New(401, "REFRESH_TOKEN_INVALID", "Refresh token tidak cocok", "")
		}
	}

	userSession, err := s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
	if err != nil {
		return "", "", apperror.New(401, "USER_SESSION_NOT_FOUND", "Sesi user tidak ditemukan", err.Error())
	}

	if !userSession.IsActive {
		return "", "", apperror.New(401, "SESSION_INACTIVE", "Sesi sudah tidak aktif", "")
	}

	if time.Now().Unix() > userSession.ExpiresAt {
		return "", "", apperror.New(401, "SESSION_EXPIRED", "Sesi sudah kedaluwarsa", "")
	}

	if userSession.HashedRefreshToken != hashedRefreshToken {
		return "", "", apperror.New(401, "REFRESH_TOKEN_INVALID", "Refresh token tidak cocok", "")
	}

	if userSession.UserID != userID {
		return "", "", apperror.New(401, "SESSION_USER_MISMATCH", "Sesi tidak sesuai dengan user", "")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", apperror.New(404, "USER_NOT_FOUND", "User tidak ditemukan", err.Error())
		}
		return "", "", apperror.New(500, "USER_FETCH_FAILED", "Gagal mengambil data user", err.Error())
	}

	refreshDuration := getRefreshTokenDuration()
	newExpiresAt := time.Now().Add(refreshDuration).Unix()

	newRefreshTokenID := uuid.New().String()
	newRefreshToken := tokenutils.GenerateRefreshToken(userID, newRefreshTokenID)
	newHashedRefreshToken := tokenutils.HashSHA256String(newRefreshToken)

	oldRefreshKey := refreshKey
	newRefreshKey := fmt.Sprintf("refresh_token:%s", newRefreshTokenID)
	sessionKey := fmt.Sprintf("session:%d", userSession.ID)

	userSession.RefreshTokenID = newRefreshTokenID
	userSession.HashedRefreshToken = newHashedRefreshToken
	userSession.ExpiresAt = newExpiresAt

	tx := s.db.Begin()
	if tx.Error != nil {
		return "", "", apperror.New(500, "TRANSACTION_START_FAILED", "Gagal memulai transaksi", tx.Error.Error())
	}

	if err := s.userSessionRepo.UpdateTX(ctx, tx, userSession); err != nil {
		tx.Rollback()
		return "", "", apperror.New(500, "SESSION_UPDATE_FAILED", "Gagal memperbarui sesi", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return "", "", apperror.New(500, "TRANSACTION_COMMIT_FAILED", "Gagal commit transaksi", err.Error())
	}

	if err := s.cacheRepo.Set(ctx, newRefreshKey, newHashedRefreshToken, refreshDuration); err != nil {
		logger.Warn("Failed to save new refresh token to Redis",
			zap.String("refresh_token_id", newRefreshTokenID),
			zap.Error(err),
		)
	}

	if err := s.cacheRepo.Set(ctx, sessionKey, userID, refreshDuration); err != nil {
		logger.Warn("Failed to save session to Redis",
			zap.Uint("session_id", userSession.ID),
			zap.Error(err),
		)
	}

	if err := s.cacheRepo.Del(ctx, oldRefreshKey); err != nil {
		logger.Warn("Failed to delete old refresh token from Redis",
			zap.String("refresh_token_id", refreshTokenID),
			zap.Error(err),
		)
	}

	accessToken := tokenutils.GenerateAccessToken(
		userID,
		userSession.ID,
		user.Email,
		user.Username,
		user.FullName,
	)

	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	refreshTokenClaims, err := tokenutils.ValidateRefreshToken(refreshToken)
	if err != nil {
		logger.Error("Failed to validate refresh token", zap.Error(err))
		return apperror.New(401, "INVALID_REFRESH_TOKEN", "Refresh token tidak valid", err.Error())
	}

	refreshTokenID, ok := refreshTokenClaims["refresh_token_id"].(string)
	if !ok || refreshTokenID == "" {
		return apperror.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Claim refresh_token_id tidak valid", "")
	}

	userIDFloat, ok := refreshTokenClaims["user_id"].(float64)
	if !ok {
		return apperror.New(401, "INVALID_REFRESH_TOKEN_CLAIMS", "Claim user_id tidak valid", "")
	}
	userID := uint(userIDFloat)

	userSession, err := s.userSessionRepo.GetByRefreshTokenID(ctx, refreshTokenID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(401, "USER_SESSION_NOT_FOUND", "Sesi user tidak ditemukan", err.Error())
		}
		return apperror.New(500, "USER_SESSION_FETCH_FAILED", "Gagal mengambil sesi user", err.Error())
	}

	if userSession.UserID != userID {
		return apperror.New(401, "SESSION_USER_MISMATCH", "Sesi tidak sesuai dengan user", "")
	}

	if !userSession.IsActive {
		refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
		sessionKey := fmt.Sprintf("session:%d", userSession.ID)
		userSessionKey := fmt.Sprintf("user_session:%d", userID)

		if err := s.cacheRepo.Del(ctx, refreshKey); err != nil {
			logger.Warn("Failed to delete refresh token from Redis", zap.Error(err))
		}
		if err := s.cacheRepo.Del(ctx, sessionKey); err != nil {
			logger.Warn("Failed to delete session data from Redis", zap.Error(err))
		}
		if err := s.cacheRepo.SRem(ctx, userSessionKey, userSession.ID); err != nil {
			logger.Warn("Failed to remove user session ID from Redis set", zap.Error(err))
		}

		return nil
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return apperror.New(500, "TRANSACTION_START_FAILED", "Gagal memulai transaksi", tx.Error.Error())
	}

	userSession.IsActive = false
	if err := s.userSessionRepo.UpdateTX(ctx, tx, userSession); err != nil {
		tx.Rollback()
		return apperror.New(500, "USER_SESSION_UPDATE_FAILED", "Gagal memperbarui sesi user", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return apperror.New(500, "TRANSACTION_COMMIT_FAILED", "Gagal commit transaksi", err.Error())
	}

	refreshKey := fmt.Sprintf("refresh_token:%s", refreshTokenID)
	sessionKey := fmt.Sprintf("session:%d", userSession.ID)
	userSessionKey := fmt.Sprintf("user_session:%d", userID)

	if err := s.cacheRepo.Del(ctx, refreshKey); err != nil {
		logger.Warn("Failed to delete refresh token from Redis", zap.Error(err))
	}

	if err := s.cacheRepo.Del(ctx, sessionKey); err != nil {
		logger.Warn("Failed to delete session data from Redis", zap.Error(err))
	}

	if err := s.cacheRepo.SRem(ctx, userSessionKey, userSession.ID); err != nil {
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

func (s *AuthService) ForgotPasswordEmailVerification(ctx context.Context, req dto.ForgotPasswordEmailVerificationRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(404, "USER_NOT_FOUND", "User tidak ditemukan", err.Error())
		}
		return apperror.New(500, "USER_FETCH_FAILED", "Gagal mengambil data user", err.Error())
	}

	if user != nil {
		existingRedisKey := fmt.Sprintf("forgot_password:%s", req.Email)
		remainingTime, err := s.cacheRepo.TTL(ctx, existingRedisKey)
		if err == nil && remainingTime > 0 {
			return apperror.New(
				429,
				"TOO_MANY_REQUESTS",
				fmt.Sprintf("Anda sudah melakukan permintaan sebelumnya, silahkan cek email anda atau coba lagi dalam %d detik", int(remainingTime.Seconds())),
				"",
			)
		}

		verificationCode, err := tokenutils.GenerateRandomCode(200)
		if err != nil {
			logger.Error("Failed to generate verification code", zap.Error(err))
			return apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode verifikasi", err.Error())
		}

		redisKey := fmt.Sprintf("forgot_password:%s", req.Email)
		err = s.cacheRepo.Set(ctx, redisKey, verificationCode, 300*time.Second)
		if err != nil {
			logger.Error("Failed to save verification code to Redis", zap.Error(err))
			return apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode verifikasi", err.Error())
		}

		verificationLink := fmt.Sprintf("%s/auth/forgot-password/verification?code=%s&email=%s", env.ClientURL(), verificationCode, req.Email)
		go util.SendPasswordResetEmail(req.Email, req.Email, verificationLink)
		return nil
	}
	return apperror.New(404, "USER_NOT_FOUND", "User tidak ditemukan", "")
}

func (s *AuthService) SendRegistrationVerificationEmail(ctx context.Context, user *model.User) error {
	randomCode1, err := tokenutils.GenerateRandomCode(150)
	if err != nil {
		logger.Error("Failed to generate random code 1", zap.Error(err))
		return apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode acak", err.Error())
	}
	randomCode2, err := tokenutils.GenerateRandomCode(150)
	if err != nil {
		logger.Error("Failed to generate random code 2", zap.Error(err))
		return apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode acak", err.Error())
	}

	verificationLink := fmt.Sprintf("%s/auth/verification?code1=%s&userId=%d&code2=%s", env.ClientURL(), randomCode1, user.ID, randomCode2)

	linkData := map[string]string{
		"link1": randomCode1,
		"link2": randomCode2,
	}
	linkJSON, err := json.Marshal(linkData)
	if err != nil {
		logger.Error("Failed to marshal verification link data", zap.Error(err))
		return apperror.New(500, "MARSHAL_FAILED", "Gagal menyimpan kode verifikasi", err.Error())
	}

	redisKey := fmt.Sprintf("link:%d", user.ID)
	err = s.cacheRepo.Set(ctx, redisKey, linkJSON, 300*time.Second)
	if err != nil {
		logger.Error("Failed to save verification link to Redis", zap.Error(err))
		return apperror.New(500, "REDIS_SAVE_FAILED", "Gagal menyimpan kode verifikasi ke Redis", err.Error())
	}

	go util.SendVerificationEmail(user.Email, user.Username, verificationLink)

	return nil
}

func (s *AuthService) VerifyRegistrationCode(ctx context.Context, code1, code2 string, userID uint) (*model.User, error) {
	redisKey := fmt.Sprintf("link:%d", userID)
	linkData, err := s.cacheRepo.Get(ctx, redisKey)
	if err != nil {
		logger.Error("Failed to get verification link from Redis", zap.Error(err))
		return nil, apperror.New(500, "REDIS_GET_FAILED", "Gagal mendapatkan kode verifikasi", err.Error())
	}

	var link map[string]string
	if err := json.Unmarshal([]byte(linkData), &link); err != nil {
		logger.Error("Failed to unmarshal verification link data", zap.Error(err))
		return nil, apperror.New(500, "UNMARSHAL_FAILED", "Gagal memproses link verifikasi", err.Error())
	}

	if link["link1"] != code1 || link["link2"] != code2 {
		return nil, apperror.New(400, "INVALID_CODE", "Link verifikasi tidak valid", "")
	}

	user, err := s.VerifyUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.cacheRepo.Del(ctx, redisKey); err != nil {
		logger.Warn("Failed to delete verification link from Redis", zap.Error(err))
	}

	return user, nil
}

func (s *AuthService) VerifyForgotPasswordCode(ctx context.Context, code, email string) error {
	redisKey := fmt.Sprintf("forgot_password:%s", email)
	storedCode, err := s.cacheRepo.Get(ctx, redisKey)
	if err != nil {
		logger.Error("Failed to get verification code from Redis", zap.Error(err))
		return apperror.New(500, "REDIS_GET_FAILED", "Gagal mendapatkan kode verifikasi", err.Error())
	}

	if storedCode != code {
		return apperror.New(400, "INVALID_CODE", "Link verifikasi tidak valid", "")
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword string) error {
	user, err := s.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Error("Failed to get user by email", zap.Error(err))
		return err
	}
	if user == nil {
		return apperror.New(404, "USER_NOT_FOUND", "Pengguna tidak ditemukan", "")
	}

	hashedPassword, err := tokenutils.HashString(newPassword)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return apperror.New(500, "PASSWORD_HASH_FAILED", "Gagal mengenkripsi kata sandi baru", err.Error())
	}

	user.Password = &hashedPassword

	_, err = s.UpdateUserByEmail(ctx, email, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, oauthEmail, oauthFullName, oauthGivenName, oauthName, oauthProviderID, provider, userIP, userAgent string) (string, string, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("Processing OAuth callback",
		zap.String("request_id", requestID),
		zap.String("provider", provider),
		zap.String("email", oauthEmail),
	)

	existingUser, err := s.GetUserByEmail(ctx, oauthEmail)
	if err != nil {
		logger.Error("Error retrieving user by email", zap.String("request_id", requestID), zap.Error(err))
		return "", "", apperror.New(500, "USER_FETCH_FAILED", "Gagal mengambil data pengguna", err.Error())
	}

	randomCode, err := tokenutils.GenerateRandomCode(5)
	if err != nil {
		logger.Error("Error generating random code", zap.String("request_id", requestID), zap.Error(err))
		return "", "", apperror.New(500, "CODE_GENERATION_FAILED", "Gagal membuat kode acak", err.Error())
	}

	if existingUser == nil {
		nickName := oauthGivenName
		if nickName == "" {
			nickName = oauthName
		}

		newUser := dto.RegisterRequest{
			Username:   fmt.Sprintf("%s_%s", nickName, randomCode),
			Email:      oauthEmail,
			FullName:   oauthFullName,
			Provider:   strings.ToUpper(provider),
			ProviderID: &oauthProviderID,
		}
		createdUser, err := s.Register(ctx, newUser, true)
		if err != nil {
			logger.Error("Error registering new user from OAuth", zap.String("request_id", requestID), zap.String("provider", provider), zap.Error(err))
			return "", "", apperror.New(500, "USER_REGISTRATION_FAILED", "Gagal mendaftarkan pengguna baru", err.Error())
		}
		logger.Info("New user registered via OAuth", zap.String("request_id", requestID), zap.String("provider", provider), zap.Uint("user_id", createdUser.ID))
		existingUser = createdUser
	}

	loginReq := dto.LoginRequest{
		Email:     existingUser.Email,
		Password:  "",
		IPAddress: userIP,
		UserAgent: userAgent,
		Provider:  strings.ToUpper(provider),
	}

	_, accessToken, refreshToken, err := s.Login(ctx, loginReq)
	if err != nil {
		logger.Error("Login failed for OAuth user", zap.String("request_id", requestID), zap.String("provider", provider), zap.Error(err))
		return "", "", apperror.New(500, "LOGIN_FAILED", "Gagal masuk dengan akun OAuth", err.Error())
	}

	logger.Info("OAuth user logged in successfully", zap.String("request_id", requestID), zap.String("provider", provider), zap.Uint("user_id", existingUser.ID))

	return accessToken, refreshToken, nil
}
