package service

import (
	"context"
	"errors"
	"pingspot/internal/domain/user_service/dto"
	"pingspot/internal/domain/user_service/repository"
	"pingspot/internal/model"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"
	contextutils "pingspot/pkg/utils/context_util"
	tokenutils "pingspot/pkg/utils/token_util"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
	followRepo      repository.FollowRepository
	db              *gorm.DB
}

func NewUserService(db *gorm.DB, userRepo repository.UserRepository, userProfileRepo repository.UserProfileRepository, followRepo repository.FollowRepository) *UserService {
	return &UserService{
		db:              db,
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
		followRepo:      followRepo,
	}
}

func (s *UserService) SaveProfile(ctx context.Context, userID uint, req dto.SaveUserProfileRequest) (*dto.SaveUserProfileResponse, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("Saving user profile",
		zap.String("request_id", requestID),
		zap.Uint("user_id", userID),
	)

	tx := s.db.Begin()
	if tx.Error != nil {
		logger.Error("Failed to start transaction",
			zap.String("request_id", requestID),
			zap.Error(tx.Error),
		)
		return nil, apperror.New(500, "TRANSACTION_START_FAILED", "gagal memulai transaksi", tx.Error.Error(), nil)
	}

	currentUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "", nil)
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mengambil data pengguna", err.Error(), nil)
	}

	if req.Username != nil && *req.Username != currentUser.Username {
		_, err := s.userRepo.GetByUsername(ctx, *req.Username)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				return nil, apperror.New(500, "USERNAME_CHECK_FAILED", "gagal memeriksa keberadaan username", err.Error(), nil)
			}
		} else {
			tx.Rollback()
			return nil, apperror.New(409, "USERNAME_EXISTS", "username sudah digunakan. Silakan pilih username lain.", "", nil)
		}
		currentUser.IsDefaultUsername = false
	}

	currentUser.FullName = req.FullName
	currentUser.Username = *req.Username

	updatedUser, err := s.userRepo.UpdateTX(ctx, tx, currentUser)
	if err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "USER_UPDATE_FAILED", "gagal memperbarui data pengguna", err.Error(), nil)
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
				return nil, apperror.New(500, "PROFILE_CREATE_FAILED", "gagal membuat profil", err.Error(), nil)
			}
			if err := tx.Commit().Error; err != nil {
				return nil, apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyimpan perubahan", err.Error(), nil)
			}
			newProfileResponse := dto.SaveUserProfileResponse{
				UserID:         userID,
				Bio:            req.Bio,
				Username:       updatedUser.Username,
				ProfilePicture: req.ProfilePicture,
				Birthday:       req.Birthday,
				Gender:         req.Gender,
				FullName:       updatedUser.FullName,
			}
			logger.Info("User profile created successfully",
				zap.String("request_id", requestID),
				zap.Uint("user_id", userID),
			)
			return &newProfileResponse, nil
		} else {
			tx.Rollback()
			return nil, apperror.New(500, "PROFILE_FETCH_FAILED", "gagal mengambil profil", err.Error(), nil)
		}
	}

	profile.Bio = req.Bio
	profile.ProfilePicture = req.ProfilePicture
	profile.Birthday = req.Birthday
	profile.Gender = req.Gender

	if _, err := s.userProfileRepo.UpdateTX(ctx, tx, profile); err != nil {
		tx.Rollback()
		return nil, apperror.New(500, "PROFILE_UPDATE_FAILED", "gagal memperbarui profil", err.Error(), nil)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, apperror.New(500, "TRANSACTION_COMMIT_FAILED", "gagal menyimpan perubahan", err.Error(), nil)
	}

	profileResponse := dto.SaveUserProfileResponse{
		UserID:         userID,
		Bio:            profile.Bio,
		ProfilePicture: profile.ProfilePicture,
		Birthday:       profile.Birthday,
		Gender:         profile.Gender,
		FullName:       updatedUser.FullName,
		Username:       updatedUser.Username,
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
		return nil, apperror.New(500, "USER_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengguna", err.Error(), nil)
	}

	usersByGender, err := s.userRepo.GetByUserGenderCount(ctx)
	if err != nil {
		return nil, apperror.New(500, "USER_GENDER_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengguna berdasarkan gender", err.Error(), nil)
	}
	totalKnownGender := usersByGender["male"] + usersByGender["female"]

	if totalKnownGender < totalUsers {
		usersByGender["unknown"] = totalUsers - totalKnownGender
	}

	monthlyUserCounts, err := s.userRepo.GetMonthlyUserCounts(ctx)
	if err != nil {
		return nil, apperror.New(500, "MONTHLY_USER_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengguna bulanan", err.Error(), nil)
	}

	return &dto.GetUserStatisticsResponse{
		TotalUsers:        totalUsers,
		UsersByGender:     usersByGender,
		MonthlyUserCounts: monthlyUserCounts,
	}, nil
}

func (s *UserService) SearchUsers(ctx context.Context, searchQuery string, usersDataNextCursor uint, limit int) (*dto.SearchResponse, error) {
	requestID := contextutils.GetRequestID(ctx)
	logger.Info("Performing search",
		zap.String("request_id", requestID),
		zap.String("search_query", searchQuery),
		zap.Int("limit", limit),
	)

	usersData, err := s.userRepo.FullTextSearchUsersPaginated(ctx, strings.ToLower(searchQuery), limit, usersDataNextCursor)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Failed to search users",
			zap.String("request_id", requestID),
			zap.String("search_query", searchQuery),
			zap.Error(err),
		)
		return nil, apperror.New(500, "USER_SEARCH_FAILED", "Gagal mencari data pengguna", err.Error(), nil)
	}

	resultUsers := make([]dto.SearchUsers, 0, len(*usersData))

	for _, user := range *usersData {
		userDTO := dto.SearchUsers{
			UserID:         user.ID,
			FullName:       user.FullName,
			Bio:            user.Profile.Bio,
			ProfilePicture: user.Profile.ProfilePicture,
			Username:	   user.Username,
			Birthday:   	   user.Profile.Birthday,
		}
		resultUsers = append(resultUsers, userDTO)
	}

	logger.Info("Search completed successfully",
		zap.String("request_id", requestID),
		zap.Int("users_found", len(*usersData)),
	)

	return &dto.SearchResponse{
		UsersData: resultUsers,
	}, nil
}


func (s *UserService) GetProfileByUsername(ctx context.Context, username string) (*dto.GetProfileResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "", nil)
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan profil user", err.Error(), nil)
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

func (s *UserService) GetProfile(ctx context.Context, userID uint) (*dto.GetProfileResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "", nil)
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mendapatkan profil user", err.Error(), nil)
	}

	missingFields := []string{}

	if user.FullName == "" {
		missingFields = append(missingFields, "fullName")
	}
	if user.Username == "" {
		missingFields = append(missingFields, "username")
	}
	if user.Profile.Bio == nil {
		missingFields = append(missingFields, "bio")
	}
	if user.Profile.ProfilePicture == nil {
		missingFields = append(missingFields, "profilePicture")
	}
	if user.Profile.Birthday == nil {
		missingFields = append(missingFields, "birthday")
	}
	if user.Profile.Gender == nil {
		missingFields = append(missingFields, "gender")
	}

	isCompleteProfile := len(missingFields) == 0

	return &dto.GetProfileResponse{
		UserID:         user.ID,
		FullName:       user.FullName,
		Bio:            user.Profile.Bio,
		ProfilePicture: user.Profile.ProfilePicture,
		Username:       user.Username,
		Birthday:       user.Profile.Birthday,
		Gender:         user.Profile.Gender,
		Email:          user.Email,
		IsCompleteProfile: isCompleteProfile,
		MissingFields:     missingFields,
		IsDefaultUsername: user.IsDefaultUsername,
	}, nil
}

func (s *UserService) SaveSecurity(ctx context.Context, userID uint, req dto.SaveUserSecurityRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "", nil)
		}
		return apperror.New(500, "USER_FETCH_FAILED", "gagal mengambil data pengguna", err.Error(), nil)
	}

	isValidPassword := false
	if user.Password != nil {
		isValidPassword = tokenutils.CheckHashString(req.CurrentPassword, *user.Password)
	}

	if !isValidPassword {
		return apperror.New(400, "INVALID_PASSWORD", "Kata sandi lama anda salah", "", nil)
	}

	hashedPassword, err := tokenutils.HashString(req.NewPassword)
	if err != nil {
		return apperror.New(500, "PASSWORD_HASH_FAILED", "Gagal mengenkripsi kata sandi", "", nil)
	}

	user.Password = &hashedPassword
	if err := s.userRepo.Save(ctx, user); err != nil {
		return apperror.New(500, "PASSWORD_UPDATE_FAILED", "Gagal memperbarui kata sandi", err.Error(), nil)
	}

	return nil
}

func (s *UserService) Follow(ctx context.Context, userID uint, req dto.FollowRequest) (*dto.FollowResponse, error) {
	currentUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.New(404, "USER_NOT_FOUND", "pengguna tidak ditemukan", "", nil)
		}
		return nil, apperror.New(500, "USER_FETCH_FAILED", "gagal mengambil data pengguna", err.Error(), nil)
	}

	var followProcess string
	
	follow := model.Follow{
		FollowingID:   req.FollowingID,
		FollowingType: model.FollowingType(req.FollowingType),
		FollowerUserID:  currentUser.ID,
	}

	existingFollow, err := s.followRepo.GetByFollowerAndFollowing(ctx, currentUser.ID, req.FollowingID, follow.FollowingType)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.New(500, "FOLLOW_CHECK_FAILED", "gagal memeriksa status mengikuti", err.Error(), nil)
	}

	if existingFollow != nil {
		unfollowErr := s.followRepo.DeleteTX(ctx, s.db, existingFollow)
		if unfollowErr != nil {
			return nil, apperror.New(500, "UNFOLLOW_FAILED", "gagal berhenti mengikuti pengguna", unfollowErr.Error(), nil)
		}
		followProcess = "unfollow"
	} else {
		if err := s.followRepo.CreateTX(ctx, s.db, &follow); err != nil {
			return nil, apperror.New(500, "FOLLOW_FAILED", "gagal mengikuti pengguna", err.Error(), nil)
		}
		followProcess = "follow"
	}


	return &dto.FollowResponse{
		FollowingID:   follow.FollowingID,
		FollowingType: string(follow.FollowingType),
		FollowerUserID:  follow.FollowerUserID,
		FollowProcess: followProcess,
	}, nil
}