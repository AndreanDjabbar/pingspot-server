package service

import (
	"context"
	"errors"
	"pingspot/internal/domain/user_service/dto"
	userRepository "pingspot/internal/domain/user_service/repository"
	socialRepository "pingspot/internal/domain/social_service/repository"
	"pingspot/internal/model"
	apperror "pingspot/pkg/app_error"
	"gorm.io/gorm"
)

type SocialService struct {
	followRepo      socialRepository.FollowRepository
	userRepo        userRepository.UserRepository
	db              *gorm.DB
}

func NewSocialService(db *gorm.DB, followRepo socialRepository.FollowRepository, userRepo userRepository.UserRepository) *SocialService {
	return &SocialService{
		db:           db,
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

func (s *SocialService) Follow(ctx context.Context, userID uint, req dto.FollowRequest) (*dto.FollowResponse, error) {
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

func (s *SocialService) GetFollowing(ctx context.Context, followingID uint, followingType string) (*dto.GetFollowDataResponse, error) {
	followingTypeEnum := model.FollowingType(followingType)
	followersCount, err := s.followRepo.GetFollowersCount(ctx, followingID, followingTypeEnum)
	if err != nil {
		return nil, apperror.New(500, "FOLLOWERS_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah pengikut", err.Error(), nil)
	}
	followingCount, err := s.followRepo.GetFollowingCount(ctx, followingID, followingTypeEnum)
	if err != nil {
		return nil, apperror.New(500, "FOLLOWING_COUNT_FETCH_FAILED", "gagal mendapatkan jumlah yang diikuti", err.Error(), nil)
	}
	return &dto.GetFollowDataResponse{
		FollowingID:   followingID,
		FollowersCount: followersCount,
		FollowingCount: followingCount,
	}, nil
}