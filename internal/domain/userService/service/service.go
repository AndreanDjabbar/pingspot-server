package service

import (
	"context"
	"errors"
	"pingspot/internal/domain/model"
	"pingspot/internal/domain/userService/dto"
	"pingspot/internal/domain/userService/repository"
	apperror "pingspot/pkg/apperror"
	contextutils "pingspot/pkg/utils/contextUtils"
	"pingspot/pkg/logger"
	tokenutils "pingspot/pkg/utils/tokenutils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
}

func NewUserService(userRepo repository.UserRepository, userProfileRepo repository.UserProfileRepository) *UserService {
	return &UserService{
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
	}
}

func (s *UserService) SaveProfile(ctx context.Context, db *gorm.DB, userID uint, req dto.SaveUserProfileRequest) (*dto.SaveUserProfileResponse, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("Saving user profile",
		zap.String("request_id", requestID),
		zap.Uint("user_id", userID),
	)

	tx := db.Begin()
	if tx.Error != nil {
		logger.Error("Failed to start transaction",
			zap.String("request_id", requestID),
			zap.Error(tx.Error),
		)
		return nil, apperror.New(500, "TRANSACTION_START_FAILED", "gagal memulai transaksi", tx.Error.Error())
	}

	if err := s.userRepo.UpdateFullNameTX(ctx, tx, userID, req.FullName); err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "FULLNAME_UPDATE_FAILED", "gagal memperbarui nama lengkap", err.Error())
	}

	profile, err := s.userProfileRepo.GetByIDTX(ctx, tx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newProfile := model.UserProfile{
				UserID:         userID,
				Bio:            req.Bio,
				ProfilePicture: req.ProfilePicture,
				Birthday:       req.Birthday,
				Gender:         req.Gender,
			}
			if _, err := s.userProfileRepo.CreateTX(ctx, tx, &newProfile); err != nil {
				tx.Rollback()
				return nil, apperror.New(500, "PROFILE_CREATE_FAILED", "gagal membuat profil", err.Error())
			}
			if err := tx.Commit().Error; err != nil {
				return nil, apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyimpan perubahan", err.Error())
			}
			newProfileResponse := dto.SaveUserProfileResponse{
				UserID:         userID,
				Bio:            req.Bio,
				ProfilePicture: req.ProfilePicture,
				Birthday:       req.Birthday,
				Gender:         req.Gender,
				FullName:       req.FullName,
			}
			logger.Info("User profile created successfully",
				zap.String("request_id", requestID),
				zap.Uint("user_id", userID),
			)
			return &newProfileResponse, nil
		} else {
			tx.Rollback()
			return nil, apperror.New(500, "PROFILE_FETCH_FAILED", "gagal mengambil profil", err.Error())
		}
	}

	profile.Bio = req.Bio
	profile.ProfilePicture = req.ProfilePicture
	profile.Birthday = req.Birthday
	profile.Gender = req.Gender

	if _, err := s.userProfileRepo.UpdateTX(ctx, tx, profile); err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "PROFILE_UPDATE_FAILED", "gagal memperbarui profil", err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyimpan perubahan", err.Error())
	}

	profileResponse := dto.SaveUserProfileResponse{
		UserID:         userID,
		Bio:            profile.Bio,
		ProfilePicture: profile.ProfilePicture,
		Birthday:       profile.Birthday,
		Gender:         profile.Gender,
		FullName:       req.FullName,
		Username:       *req.Username,
	}
	logger.Info("User profile updated successfully",
		zap.String("request_id", requestID),
		zap.Uint("user_id", userID),
	)
	return &profileResponse, nil
}

func (s *UserService) GetUserStatistics(ctx context.Context) (*dto.GetUserStatisticsResponse, error) {
	totalUsers, err := s.userRepo.GetUsersCount(ctx)
	if err != nil {
		return nil, apperror.New(500, "USER_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengguna", err.Error())
	}

	usersByGender, err := s.userRepo.GetByUserGenderCount(ctx)
	if err != nil {
		return nil, apperror.New(500, "USER_GENDER_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengguna berdasarkan gender", err.Error())
	}
	totalKnownGender := usersByGender["male"] + usersByGender["female"]

	if totalKnownGender < totalUsers {
		usersByGender["unknown"] = totalUsers - totalKnownGender
	}

	monthlyUserCounts, err := s.userRepo.GetMonthlyUserCounts(ctx)
	if err != nil {
		return nil, apperror.New(500, "MONTHLY_USER_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengguna bulanan", err.Error())
	}

	return &dto.GetUserStatisticsResponse{
		TotalUsers:        totalUsers,
		UsersByGender:     usersByGender,
		MonthlyUserCounts: monthlyUserCounts,
	}, nil
}

func (s *UserService) GetProfileByUsername(ctx context.Context, username string) (*dto.GetProfileResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "")
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan profil user", err.Error())
	}
	return &dto.GetProfileResponse{
		UserID:         user.ID,
		FullName:       user.FullName,
		Bio:            user.Profile.Bio,
		ProfilePicture: user.Profile.ProfilePicture,
		Username:       user.Username,
		Birthday:       user.Profile.Birthday,
		Gender:		 	user.Profile.Gender,
		Email:          user.Email,
	}, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*dto.GetProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "")
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan profil user", err.Error())
	}

	return &dto.GetProfileResponse{
		UserID:         user.ID,
		FullName:       user.FullName,
		Bio:            user.Profile.Bio,
		ProfilePicture: user.Profile.ProfilePicture,
		Username:       user.Username,
		Birthday:       user.Profile.Birthday,
		Gender:         user.Profile.Gender,
		Email:          user.Email,
	}, nil
}

func (s *UserService) SaveSecurity(ctx context.Context, userID uint, req dto.SaveUserSecurityRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "")
		}
		return apperror.New(500, "USER_FETCH_FAILED", "gagal mengambil data pengguna", err.Error())
	}

	isValidPassword := false
	if user.Password != nil {
		isValidPassword = tokenutils.CheckHashString(req.CurrentPassword, *user.Password)
	}

	if !isValidPassword {
		return apperror.New(400, "INVALID_PASSWORD", "Kata sandi lama anda salah", "")
	}

	hashedPassword, err := tokenutils.HashString(req.NewPassword)
	if err != nil {
		return apperror.New(500, "PASSWORD_HASH_FAILED", "Gagal mengenkripsi kata sandi", "")
	}

	user.Password = &hashedPassword
	if err := s.userRepo.Save(ctx, user); err != nil {
		return apperror.New(500, "PASSWORD_UPDATE_FAILED", "Gagal memperbarui kata sandi", err.Error())
	}

	return nil
}
