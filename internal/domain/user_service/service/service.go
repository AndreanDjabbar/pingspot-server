package service

import (
	"context"
	"errors"
	"pingspot/internal/domain/user_service/dto"
	"pingspot/internal/domain/user_service/repository"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/logger"
	"pingspot/internal/model"
	"pingspot/internal/util"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserService struct {
	txManager	   database.TransactionManager
	userRepo        repository.UserRepository
	userProfileRepo repository.UserProfileRepository
}

func NewUserService(txManager database.TransactionManager, userRepo repository.UserRepository, userProfileRepo repository.UserProfileRepository) *UserService {
	return &UserService{
		txManager:       txManager,
		userRepo:        userRepo,
		userProfileRepo: userProfileRepo,
	}
}


func (s *UserService) SaveProfile(
    ctx context.Context,
    userID uint,
    req dto.SaveUserProfileRequest,
) (*dto.SaveUserProfileResponse, error) {
    requestID := util.GetRequestID(ctx)
    logger.Info("Saving user profile",
        zap.String("request_id", requestID),
        zap.Uint("user_id", userID),
    )

    var response dto.SaveUserProfileResponse
    var profileCreated bool

    err := s.txManager.Transaction(ctx, func(tx *gorm.DB) error {
        userRepo := s.userRepo.WithTX(tx)
        profileRepo := s.userProfileRepo.WithTX(tx)

        currentUser, err := userRepo.GetByID(ctx, userID)
        if err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return util.New(
                    404,
                    "USER_NOT_FOUND",
                    "User not found",
                    "",
                    nil,
                )
            }
            return util.New(
                500,
                "USER_FETCH_FAILED",
                "Failed to retrieve user data",
                err.Error(),
                nil,
            )
        }

        if req.Username != nil && *req.Username != currentUser.Username {
            _, err := userRepo.GetByUsername(ctx, *req.Username)
            if err == nil {
                return util.New(
                    409,
                    "USERNAME_EXISTS",
                    "Username is already in use. Please choose another username",
                    "",
                    nil,
                )
            }
            if !errors.Is(err, gorm.ErrRecordNotFound) {
                return util.New(
                    500,
                    "USERNAME_CHECK_FAILED",
                    "Failed to check username availability",
                    err.Error(),
                    nil,
                )
            }
            currentUser.Username = *req.Username
            currentUser.IsDefaultUsername = false
        }

        currentUser.FullName = req.FullName

        updatedUser, err := userRepo.Update(ctx, currentUser)
        if err != nil {
            return util.New(
                500,
                "USER_UPDATE_FAILED",
                "Failed to update user data",
                err.Error(),
                nil,
            )
        }
        profile, err := profileRepo.GetByID(ctx, userID)

        if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
            return util.New(
                500,
                "PROFILE_FETCH_FAILED",
                "Failed to retrieve profile",
                err.Error(),
                nil,
            )
        }
		
        if errors.Is(err, gorm.ErrRecordNotFound) {
            profile = &model.UserProfile{
                UserID:         userID,
                Bio:            req.Bio,
                ProfilePicture: req.ProfilePicture,
                Birthday:       req.Birthday,
                Gender:         req.Gender,
            }

            if _, err := profileRepo.Create(ctx, profile); err != nil {
                return util.New(
                    500,
                    "PROFILE_CREATE_FAILED",
                    "Failed to create profile",
                    err.Error(),
                    nil,
                )
            }

            profileCreated = true
        } else {
            profile.Bio = req.Bio
            profile.ProfilePicture = req.ProfilePicture
            profile.Birthday = req.Birthday
            profile.Gender = req.Gender

            if _, err := profileRepo.Update(ctx, profile); err != nil {
                return util.New(
                    500,
                    "PROFILE_UPDATE_FAILED",
                    "Failed to update profile",
                    err.Error(),
                    nil,
                )
            }
        }

        response = dto.SaveUserProfileResponse{
            UserID:         userID,
            Bio:            profile.Bio,
            ProfilePicture: profile.ProfilePicture,
            Birthday:       profile.Birthday,
            Gender:         profile.Gender,
            FullName:       updatedUser.FullName,
            Username:       updatedUser.Username,
        }
		
        return nil
    })

    if err != nil {
        logger.Error("Failed to save user profile",
            zap.String("request_id", requestID),
            zap.Uint("user_id", userID),
            zap.Error(err),
        )

        return nil, err
    }

    action := "updated"
    if profileCreated {
        action = "created"
    }

    logger.Info("User profile saved successfully",
        zap.String("request_id", requestID),
        zap.Uint("user_id", userID),
        zap.String("action", action),
    )

    return &response, nil
}


func (s *UserService) GetUserStatistics(ctx context.Context) (*dto.GetUserStatisticsResponse, error) {
	totalUsers, err := s.userRepo.GetUsersCount(ctx)
	if err != nil {
		return nil, util.New(500, "USER_COUNT_FETCH_FAILED", "Failed to retrieve user count", err.Error(), nil)
	}

	usersByGender, err := s.userRepo.GetByUserGenderCount(ctx)
	if err != nil {
		return nil, util.New(500, "USER_GENDER_COUNT_FETCH_FAILED", "Failed to retrieve user count by gender", err.Error(), nil)
	}
	totalKnownGender := usersByGender["male"] + usersByGender["female"]

	if totalKnownGender < totalUsers {
		usersByGender["unknown"] = totalUsers - totalKnownGender
	}

	monthlyUserCounts, err := s.userRepo.GetMonthlyUserCounts(ctx)
	if err != nil {
		return nil, util.New(500, "MONTHLY_USER_COUNT_FETCH_FAILED", "Failed to retrieve monthly user count", err.Error(), nil)
	}

	return &dto.GetUserStatisticsResponse{
		TotalUsers:        totalUsers,
		UsersByGender:     usersByGender,
		MonthlyUserCounts: monthlyUserCounts,
	}, nil
}

func (s *UserService) SearchUsers(ctx context.Context, searchQuery string, usersDataNextCursor uint, limit int) (*dto.SearchResponse, error) {
	requestID := util.GetRequestID(ctx)
	logger.Info("Performing search",
		zap.String("request_id", requestID),
		zap.String("search_query", searchQuery),
		zap.Int("limit", limit),
	)

	usersData, err := s.userRepo.FullTextSearchUsersPaginated(ctx, strings.ToLower(searchQuery), limit, usersDataNextCursor, true)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Failed to search users",
			zap.String("request_id", requestID),
			zap.String("search_query", searchQuery),
			zap.Error(err),
		)
		return nil, util.New(500, "USER_SEARCH_FAILED", "Failed to search user data", err.Error(), nil)
	}

	resultUsers := make([]dto.SearchUsers, 0, len(*usersData))

	for _, user := range *usersData {
		userDTO := dto.SearchUsers{
			UserID:         user.ID,
			FullName:       user.FullName,
			Bio:            user.Profile.Bio,
			ProfilePicture: user.Profile.ProfilePicture,
			Username:       user.Username,
			Birthday:       user.Profile.Birthday,
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
			return nil, util.New(404, "USER_NOT_FOUND", "User not found", "", nil)
		}
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user profile", err.Error(), nil)
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
			return nil, util.New(404, "USER_NOT_FOUND", "User not found", "", nil)
		}
		return nil, util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user profile", err.Error(), nil)
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
		UserID:                     user.ID,
		FullName:                   user.FullName,
		Bio:                        user.Profile.Bio,
		ProfilePicture:             user.Profile.ProfilePicture,
		Username:                   user.Username,
		Birthday:                   user.Profile.Birthday,
		Gender:                     user.Profile.Gender,
		Email:                      user.Email,
		IsCompleteProfile:          isCompleteProfile,
		IsDisableEmailNotification: user.IsDisableEmailNotification,
		MissingFields:              missingFields,
		IsDefaultUsername:          user.IsDefaultUsername,
	}, nil
}

func (s *UserService) SaveSecurity(ctx context.Context, userID uint, req dto.SaveUserSecurityRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.New(404, "USER_NOT_FOUND", "User not found", "", nil)
		}
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	isValidPassword := false
	if user.Password != nil {
		isValidPassword = util.CheckHashString(req.CurrentPassword, *user.Password)
	}

	if !isValidPassword {
		return util.New(400, "INVALID_PASSWORD", "Your old password is incorrect", "", nil)
	}

	hashedPassword, err := util.HashString(req.NewPassword)
	if err != nil {
		return util.New(500, "PASSWORD_HASH_FAILED", "Failed to hash password", "", nil)
	}

	user.Password = &hashedPassword
	if err := s.userRepo.Save(ctx, user); err != nil {
		return util.New(500, "PASSWORD_UPDATE_FAILED", "Failed to update password", err.Error(), nil)
	}

	return nil
}

func (s *UserService) UpdateEmailNotificationPreference(ctx context.Context, userID uint, req dto.UpdateEmailNotificationPreferenceRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return util.New(404, "USER_NOT_FOUND", "User not found", "", nil)
		}
		return util.New(500, "USER_FETCH_FAILED", "Failed to retrieve user data", err.Error(), nil)
	}

	user.IsDisableEmailNotification = *req.IsDisableEmailNotification

	if err := s.userRepo.Save(ctx, user); err != nil {
		return util.New(500, "EMAIL_NOTIFICATION_UPDATE_FAILED", "Failed to update email notification preferences", err.Error(), nil)
	}

	return nil
}
