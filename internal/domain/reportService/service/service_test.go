package service

import (
	"context"
	"errors"
	"pingspot/internal/domain/mocks"
	"pingspot/internal/domain/mocks/report"
	"pingspot/internal/domain/model"
	"pingspot/internal/domain/reportService/dto"
	mainutils "pingspot/pkg/utils/mainUtils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&model.Report{},
		&model.ReportImage{},
		&model.ReportProgress{},
		&model.ReportReaction{},
		&model.ReportVote{},
		&model.User{},
		&model.UserProfile{},
	)
	require.NoError(t, err)

	return db
}

func TestNewReportService(t *testing.T) {
	t.Run("should create new report service", func(t *testing.T) {
		mockReportRepo := new(report.MockReportRepository)
		mockReportImageRepo := new(report.MockReportImageRepository)
		mockReportProgressRepo := new(report.MockReportProgressRepository)
		mockReportCommentRepo := new(report.MockReportCommentRepository)
		mockReportVoteRepo := new(report.MockReportVoteRepository)
		mockReportLocationRepo := new(report.MockReportLocationRepository)
		mockReportReactionRepo := new(report.MockReportReactionRepository)
		mockUserRepo := new(mocks.MockUserRepository)
		mockUserProfileRepo := new(mocks.MockUserProfileRepository)
		mockTaskService := new(mocks.MockTaskService)
		service := NewreportService(
			mockReportRepo,
			mockReportLocationRepo,
			mockReportReactionRepo,
			mockReportImageRepo,
			mockUserRepo,
			mockUserProfileRepo,
			mockReportProgressRepo,
			mockReportVoteRepo,
			mockTaskService,
			mockReportCommentRepo,
		)

		require.NotNil(t, service)
	})
}

func setupMocks() (
	*report.MockReportRepository,
	*report.MockReportLocationRepository,
	*report.MockReportReactionRepository,
	*report.MockReportImageRepository,
	*mocks.MockUserRepository,
	*mocks.MockUserProfileRepository,
	*report.MockReportProgressRepository,
	*report.MockReportVoteRepository,
	*mocks.MockTaskService,
	*report.MockReportCommentRepository,
	*ReportService,
) {
	mockReportRepo := new(report.MockReportRepository)
	mockReportLocationRepo := new(report.MockReportLocationRepository)
	mockReportReactionRepo := new(report.MockReportReactionRepository)
	mockReportImageRepo := new(report.MockReportImageRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockUserProfileRepo := new(mocks.MockUserProfileRepository)
	mockReportProgressRepo := new(report.MockReportProgressRepository)
	mockReportVoteRepo := new(report.MockReportVoteRepository)
	mockTaskService := new(mocks.MockTaskService)
	mockReportCommentRepo := new(report.MockReportCommentRepository)

	service := NewreportService(
		mockReportRepo,
		mockReportLocationRepo,
		mockReportReactionRepo,
		mockReportImageRepo,
		mockUserRepo,
		mockUserProfileRepo,
		mockReportProgressRepo,
		mockReportVoteRepo,
		mockTaskService,
		mockReportCommentRepo,
	)

	return mockReportRepo, mockReportLocationRepo, mockReportReactionRepo, mockReportImageRepo,
		mockUserRepo, mockUserProfileRepo, mockReportProgressRepo, mockReportVoteRepo,
		mockTaskService, mockReportCommentRepo, service
}

func TestReportService_CreateReport(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("should create report successfully", func(t *testing.T) {
		mockReportRepo, mockReportLocationRepo, _, mockReportImageRepo, _, _, _, _, _, _, service := setupMocks()

		req := dto.CreateReportRequest{
			ReportTitle:       "Test Report",
			ReportDescription: "Test Description",
			ReportType:        "INFRASTRUCTURE",
			HasProgress:       mainutils.BoolPtrOrNil(true),
			Latitude:          -6.200000,
			Longitude:         106.816666,
			DetailLocation:    "Test Location",
			DisplayName:       mainutils.StrPtrOrNil("Test Display"),
			MapZoom:           mainutils.IntPtrOrNil(15),
			Image1URL:         mainutils.StrPtrOrNil("image1.jpg"),
		}

		mockReportRepo.On("Create", ctx, mock.AnythingOfType("*model.Report"), mock.AnythingOfType("*gorm.DB")).
			Return(nil).Run(func(args mock.Arguments) {
			report := args.Get(1).(*model.Report)
			report.ID = 1
		})
		mockReportLocationRepo.On("Create", ctx, mock.AnythingOfType("*model.ReportLocation"), mock.AnythingOfType("*gorm.DB")).Return(nil)
		mockReportImageRepo.On("Create", ctx, mock.AnythingOfType("*model.ReportImage"), mock.AnythingOfType("*gorm.DB")).Return(nil)

		result, err := service.CreateReport(ctx, db, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.ReportTitle, result.Report.ReportTitle)
		assert.Equal(t, model.WAITING, result.Report.ReportStatus)
		mockReportRepo.AssertExpectations(t)
		mockReportLocationRepo.AssertExpectations(t)
		mockReportImageRepo.AssertExpectations(t)
	})

	t.Run("should return error when report creation fails", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		req := dto.CreateReportRequest{
			ReportTitle:       "Test Report",
			ReportDescription: "Test Description",
			ReportType:        "INFRASTRUCTURE",
		}

		mockReportRepo.On("Create", ctx, mock.AnythingOfType("*model.Report"), mock.AnythingOfType("*gorm.DB")).
			Return(errors.New("database error"))

		result, err := service.CreateReport(ctx, db, 1, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Gagal")
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_EditReport(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("should edit report successfully", func(t *testing.T) {
		mockReportRepo, mockReportLocationRepo, _, mockReportImageRepo, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:                1,
			UserID:            1,
			ReportTitle:       "Old Title",
			ReportDescription: "Old Description",
			ReportStatus:      model.WAITING,
			ReportType:        model.Infrastructure,
			ReportLocation: &model.ReportLocation{
				ID:             1,
				ReportID:       1,
				Latitude:       -6.300000,
				Longitude:      106.816666,
				DetailLocation: "Old Location",
				DisplayName:    mainutils.StrPtrOrNil("Old Display"),
				MapZoom:        mainutils.IntPtrOrNil(12),
			},
			ReportImages: &model.ReportImage{
				ID:        1,
				ReportID:  1,
				Image1URL: mainutils.StrPtrOrNil("old_image1.jpg"),
			},
		}

		req := dto.EditReportRequest{
			ReportTitle:       "New Title",
			ReportDescription: "New Description",
			ReportType:        "ENVIRONMENT",
			HasProgress:       mainutils.BoolPtrOrNil(false),
		}

		mockReportRepo.On("GetByIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(existingReport, nil)
		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{}, nil)
		mockReportLocationRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportLocation")).
			Return(&model.ReportLocation{}, nil)
		mockReportImageRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportImage")).
			Return(&model.ReportImage{}, nil)

		result, err := service.EditReport(ctx, db, 1, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockReportLocationRepo.AssertExpectations(t)
		mockReportImageRepo.AssertExpectations(t)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when report not found", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		req := dto.EditReportRequest{
			ReportTitle: "New Title",
		}

		mockReportRepo.On("GetByIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(nil, gorm.ErrRecordNotFound)

		result, err := service.EditReport(ctx, db, 1, 1, req)

		assert.Error(t, err)
		assert.Nil(t, result, err.Error())
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when user is not owner", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
		}

		req := dto.EditReportRequest{
			ReportTitle: "New Title",
		}

		mockReportRepo.On("GetByIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(existingReport, nil)

		result, err := service.EditReport(ctx, db, 1, 1, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tidak memiliki izin")
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_DeleteReport(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("should soft delete report successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       1,
			ReportStatus: model.WAITING,
			IsDeleted:    mainutils.BoolPtrOrNil(false),
		}

		mockReportRepo.On("GetByIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(existingReport, nil)
		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{}, nil)

		err := service.DeleteReport(ctx, db, 1, 1, "soft")

		assert.NoError(t, err)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should permanently delete report successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       1,
			ReportStatus: model.WAITING,
			IsDeleted:    mainutils.BoolPtrOrNil(true),
		}

		mockReportRepo.On("GetByIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(existingReport, nil)
		mockReportRepo.On("DeleteTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{}, nil)

		err := service.DeleteReport(ctx, db, 1, 1, "hard")

		assert.Nil(t, err)
		assert.NoError(t, err)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when report not found", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		mockReportRepo.On("GetByIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(nil, gorm.ErrRecordNotFound)

		err := service.DeleteReport(ctx, db, 1, 1, "soft")

		assert.Error(t, err)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_GetAllReport(t *testing.T) {
	ctx := context.Background()
	t.Run("should get all reports successfully", func(t *testing.T) {
		mockReportRepo, _, mockReportReactionRepo, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		userProfile1 := &model.UserProfile{
			UserID: 1,
		}
		user1 := &model.User{
			ID:       1,
			Username: "user1",
			Profile:  *userProfile1,
		}
		userProfile2 := &model.UserProfile{
			UserID: 2,
		}
		user2 := &model.User{
			ID:       2,
			Username: "user2",
			Profile:  *userProfile2,
		}

		reports := &[]model.Report{
			{
				ID:                1,
				UserID:            1,
				ReportTitle:       "Test Report 1",
				ReportDescription: "Test Description 1",
				ReportStatus:      model.WAITING,
				HasProgress:       mainutils.BoolPtrOrNil(true),
				CreatedAt:         time.Now().Unix(),
				ReportLocation: &model.ReportLocation{
					ReportID:  1,
					Latitude:  -6.2088,
					Longitude: 106.8456,
				},
				ReportImages: &model.ReportImage{
					ReportID:  1,
					Image1URL: mainutils.StrPtrOrNil("image1.jpg"),
				},
				ReportReactions: &[]model.ReportReaction{},
				ReportProgress:  &[]model.ReportProgress{},
				ReportVotes:     &[]model.ReportVote{},
				ReportType:      model.Infrastructure,
				User:            *user1,
			},
			{
				ID:                2,
				UserID:            2,
				ReportTitle:       "Test Report 2",
				ReportDescription: "Test Description 2",
				ReportStatus:      model.WAITING,
				HasProgress:       mainutils.BoolPtrOrNil(false),
				CreatedAt:         time.Now().Unix(),
				ReportLocation: &model.ReportLocation{
					ReportID:  2,
					Latitude:  -6.2088,
					Longitude: 106.8456,
				},
				ReportImages: &model.ReportImage{
					ReportID:  2,
					Image1URL: mainutils.StrPtrOrNil("image2.jpg"),
				},
				ReportReactions: &[]model.ReportReaction{},
				ReportProgress:  &[]model.ReportProgress{},
				ReportVotes:     &[]model.ReportVote{},
				ReportType:      model.Infrastructure,
				User:            *user2,
			},
		}

		reportsCount := dto.TotalReportCount{
			TotalReports:               2,
			TotalInfrastructureReports: 2,
		}

		distance := dto.Distance{}

		mockReportRepo.On("GetByIsDeletedPaginated", ctx, uint(5), uint(0), "", "", "", "", distance, false).
			Return(reports, nil)
		mockReportRepo.On("GetByReportTypeCount", ctx).Return(&reportsCount, nil)
		mockReportReactionRepo.On("GetLikeReactionCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportReactionRepo.On("GetDislikeReactionCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportReactionRepo.On("GetLikeReactionCount", ctx, uint(2)).Return(int64(0), nil)
		mockReportReactionRepo.On("GetDislikeReactionCount", ctx, uint(2)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetResolvedVoteCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetOnProgressVoteCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetNotResolvedVoteCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetResolvedVoteCount", ctx, uint(2)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetOnProgressVoteCount", ctx, uint(2)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetNotResolvedVoteCount", ctx, uint(2)).Return(int64(0), nil)

		result, err := service.GetAllReport(ctx, 1, 0, "", "", "", "", distance)

		assert.Nil(t, err)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Reports, 2)
		assert.Equal(t, int64(2), result.TotalCounts.TotalReports)

		mockReportRepo.AssertExpectations(t)
		mockReportReactionRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should return error when fetch reports fails", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		distance := dto.Distance{}

		mockReportRepo.On("GetByIsDeletedPaginated", ctx, uint(5), uint(0), "", "", "", "", distance, false).
			Return(nil, errors.New("database error"))

		result, err := service.GetAllReport(ctx, 1, 0, "", "", "", "", distance)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_GetReportByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get report by ID successfully", func(t *testing.T) {
		mockReportRepo, _, mockReportReactionRepo, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		userProfile := &model.UserProfile{
			UserID: 1,
		}
		user := &model.User{
			ID:       1,
			Username: "testuser",
			Profile:  *userProfile,
		}

		existingReport := &model.Report{
			ID:                1,
			UserID:            1,
			ReportTitle:       "Test Report",
			ReportDescription: "Test Description",
			ReportStatus:      model.WAITING,
			HasProgress:       mainutils.BoolPtrOrNil(false),
			CreatedAt:         time.Now().Unix(),
			ReportLocation:    &model.ReportLocation{},
			ReportProgress:    &[]model.ReportProgress{},
			ReportImages:      &model.ReportImage{},
			IsDeleted:         mainutils.BoolPtrOrNil(false),
			ReportType:        model.Infrastructure,
			User:              *user,
			ReportVotes:       &[]model.ReportVote{},
			ReportReactions:   &[]model.ReportReaction{},
		}

		mockReportRepo.On("GetByIDIsDeleted", ctx, uint(1), false).Return(existingReport, nil)
		mockReportReactionRepo.On("GetByUserReportID", ctx, uint(1), uint(1)).Return(nil, gorm.ErrRecordNotFound)
		mockReportVoteRepo.On("GetByUserReportID", ctx, uint(1), uint(1)).Return(nil, gorm.ErrRecordNotFound)
		mockReportReactionRepo.On("GetLikeReactionCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportReactionRepo.On("GetDislikeReactionCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetResolvedVoteCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetOnProgressVoteCount", ctx, uint(1)).Return(int64(0), nil)
		mockReportVoteRepo.On("GetNotResolvedVoteCount", ctx, uint(1)).Return(int64(0), nil)

		result, err := service.GetReportByID(ctx, 1, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Report", result.Report.ReportTitle)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when report not found", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		mockReportRepo.On("GetByIDIsDeleted", ctx, uint(1), false).Return(nil, gorm.ErrRecordNotFound)

		result, err := service.GetReportByID(ctx, 1, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_ReactToReport(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("should create new like reaction successfully", func(t *testing.T) {
		_, _, mockReportReactionRepo, _, _, _, _, _, _, _, service := setupMocks()

		var existingReportReaction *model.ReportReaction = nil

		mockReportReactionRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(existingReportReaction, nil)

		mockReportReactionRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportReaction")).
			Return(&model.ReportReaction{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				Type:     model.Like,
			}, nil)

		result, err := service.ReactToReport(ctx, db, 1, 1, "LIKE")

		var totalLike int64 = 0
		if result.ReactionType == "LIKE" {
			totalLike = 1
		}

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), totalLike)
		mockReportReactionRepo.AssertExpectations(t)
	})

	t.Run("should delete reaction when reaction same", func(t *testing.T) {
		_, _, mockReportReactionRepo, _, _, _, _, _, _, _, service := setupMocks()

		existingReportReaction := &model.ReportReaction{
			ID:       1,
			UserID:   1,
			ReportID: 1,
			Type:     model.Like,
		}

		mockReportReactionRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(existingReportReaction, nil)
		mockReportReactionRepo.On("DeleteTX", ctx, mock.AnythingOfType("*gorm.DB"), existingReportReaction).
			Return(nil)
		result, err := service.ReactToReport(ctx, db, 1, 1, "LIKE")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockReportReactionRepo.AssertExpectations(t)
	})

	t.Run("should update reaction when reaction different", func(t *testing.T) {
		_, _, mockReportReactionRepo, _, _, _, _, _, _, _, service := setupMocks()

		existingReportReaction := &model.ReportReaction{
			ID:       1,
			UserID:   1,
			ReportID: 1,
			Type:     model.Like,
		}
		mockReportReactionRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(existingReportReaction, nil)
		mockReportReactionRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportReaction")).
			Return(&model.ReportReaction{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				Type:     model.Dislike,
			}, nil)
		result, err := service.ReactToReport(ctx, db, 1, 1, "DISLIKE")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockReportReactionRepo.AssertExpectations(t)
	})
}

func TestReportService_VoteToReport(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("should create new vote successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(nil, gorm.ErrRecordNotFound)
		mockReportVoteRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportVote")).
			Return(&model.ReportVote{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				VoteType: model.RESOLVED,
			}, nil)
		mockReportVoteRepo.On("GetReportVoteCountsTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(map[model.ReportStatus]int64{
				model.RESOLVED:     1,
				model.ON_PROGRESS:  0,
				model.NOT_RESOLVED: 0,
			}, nil)
		mockReportVoteRepo.On("GetTotalVoteCountTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(int64(1), nil)
		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{}, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should return error when voting on own report", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:     1,
			UserID: 1,
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Anda tidak dapat memberikan suara pada laporan anda sendiri")
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when report not found", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		mockReportRepo.On("GetByID", ctx, uint(999)).Return(nil, gorm.ErrRecordNotFound)

		result, err := service.VoteToReport(ctx, db, 1, 999, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Laporan tidak ditemukan")
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when voting on resolved report", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.RESOLVED,
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Anda tidak dapat memberikan suara pada laporan yang sudah selesai")
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when voting on expired report", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.EXPIRED,
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Anda tidak dapat memberikan suara pada laporan yang sudah kedaluwarsa")
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when report has no progress", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:          1,
			UserID:      2,
			HasProgress: mainutils.BoolPtrOrNil(false),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Anda tidak dapat memberikan suara pada laporan tanpa progres (informasi saja)")
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should delete vote when voting same type again", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		existingVote := &model.ReportVote{
			ID:       1,
			UserID:   1,
			ReportID: 1,
			VoteType: model.RESOLVED,
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(existingVote, nil)
		mockReportVoteRepo.On("DeleteTX", ctx, mock.AnythingOfType("*gorm.DB"), existingVote).Return(nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.NoError(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should update vote when changing vote type", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		existingVote := &model.ReportVote{
			ID:       1,
			UserID:   1,
			ReportID: 1,
			VoteType: model.ON_PROGRESS,
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(existingVote, nil)
		mockReportVoteRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportVote")).
			Return(&model.ReportVote{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				VoteType: model.RESOLVED,
			}, nil)
		mockReportVoteRepo.On("GetReportVoteCountsTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(map[model.ReportStatus]int64{
				model.RESOLVED:     1,
				model.ON_PROGRESS:  0,
				model.NOT_RESOLVED: 0,
			}, nil)
		mockReportVoteRepo.On("GetTotalVoteCountTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(int64(1), nil)
		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{}, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, model.RESOLVED, result.VoteType)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should change status to NOT_RESOLVED when margin >= 20% and top vote is NOT_RESOLVED", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(nil, gorm.ErrRecordNotFound)
		mockReportVoteRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportVote")).
			Return(&model.ReportVote{
				ID:        1,
				UserID:    1,
				ReportID:  1,
				VoteType:  model.NOT_RESOLVED,
				CreatedAt: time.Now().Unix(),
				UpdatedAt: time.Now().Unix(),
			}, nil)
		mockReportVoteRepo.On("GetReportVoteCountsTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(map[model.ReportStatus]int64{
				model.NOT_RESOLVED: 4,
				model.ON_PROGRESS:  1,
				model.RESOLVED:     0,
			}, nil)
		mockReportVoteRepo.On("GetTotalVoteCountTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(int64(5), nil)
		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{
				ID:           1,
				ReportStatus: model.NOT_RESOLVED,
			}, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "NOT_RESOLVED")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, model.NOT_RESOLVED, result.ReportStatus)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should change status to ON_PROGRESS when margin >= 20% and top vote is ON_PROGRESS", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(nil, gorm.ErrRecordNotFound)
		mockReportVoteRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportVote")).
			Return(&model.ReportVote{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				VoteType: model.ON_PROGRESS,
			}, nil)
		mockReportVoteRepo.On("GetReportVoteCountsTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(map[model.ReportStatus]int64{
				model.ON_PROGRESS:  4,
				model.RESOLVED:     1,
				model.NOT_RESOLVED: 0,
			}, nil)
		mockReportVoteRepo.On("GetTotalVoteCountTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).Return(int64(5), nil)
		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.MatchedBy(func(r *model.Report) bool {
			return r.ReportStatus == model.ON_PROGRESS
		})).Return(&model.Report{}, nil)

		result, err := service.VoteToReport(ctx, db, 1, 1, "ON_PROGRESS")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should return error when GetReportVoteCountsTX fails", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(nil, gorm.ErrRecordNotFound)
		mockReportVoteRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportVote")).
			Return(&model.ReportVote{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				VoteType: model.RESOLVED,
			}, nil)
		mockReportVoteRepo.On("GetReportVoteCountsTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(nil, errors.New("database error"))

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})

	t.Run("should return error when GetTotalVoteCountTX fails", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, mockReportVoteRepo, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       2,
			ReportStatus: model.WAITING,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockReportVoteRepo.On("GetByUserReportIDTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1), uint(1)).
			Return(nil, gorm.ErrRecordNotFound)
		mockReportVoteRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportVote")).
			Return(&model.ReportVote{
				ID:       1,
				UserID:   1,
				ReportID: 1,
				VoteType: model.RESOLVED,
			}, nil)
		mockReportVoteRepo.On("GetReportVoteCountsTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(map[model.ReportStatus]int64{
				model.RESOLVED:     1,
				model.ON_PROGRESS:  0,
				model.NOT_RESOLVED: 0,
			}, nil)
		mockReportVoteRepo.On("GetTotalVoteCountTX", ctx, mock.AnythingOfType("*gorm.DB"), uint(1)).
			Return(int64(0), errors.New("database error"))

		result, err := service.VoteToReport(ctx, db, 1, 1, "RESOLVED")

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportVoteRepo.AssertExpectations(t)
	})
}

func TestReportService_UploadProgressReport(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB(t)

	t.Run("should upload progress successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, mockReportProgressRepo, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       1,
			ReportStatus: model.ON_PROGRESS,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		req := dto.UploadProgressReportRequest{
			Status:      "Progress update",
			Attachment1: mainutils.StrPtrOrNil("progress1.jpg"),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		mockReportProgressRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportProgress")).
			Return(&model.ReportProgress{ID: 1}, nil)

		mockReportRepo.On("UpdateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.Report")).
			Return(&model.Report{}, nil)

		result, err := service.UploadProgressReport(ctx, db, 1, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportProgressRepo.AssertExpectations(t)
	})

	t.Run("should return error when report has no progress enabled", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:          1,
			UserID:      1,
			HasProgress: mainutils.BoolPtrOrNil(false),
		}

		req := dto.UploadProgressReportRequest{
			Status: "Progress update",
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.UploadProgressReport(ctx, db, 1, 1, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when user is not the report owner", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:          1,
			UserID:      1,
			HasProgress: mainutils.BoolPtrOrNil(false),
		}

		currentUserID := uint(2)

		req := dto.UploadProgressReportRequest{
			Status: "Progress update",
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.UploadProgressReport(ctx, db, currentUserID, 1, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when report is resolved", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       1,
			ReportStatus: model.RESOLVED,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		req := dto.UploadProgressReportRequest{
			Status: "Progress update",
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		result, err := service.UploadProgressReport(ctx, db, 1, 1, req)
		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when creating progress fails", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, mockReportProgressRepo, _, _, _, service := setupMocks()

		existingReport := &model.Report{
			ID:           1,
			UserID:       1,
			ReportStatus: model.ON_PROGRESS,
			HasProgress:  mainutils.BoolPtrOrNil(true),
		}

		req := dto.UploadProgressReportRequest{
			Status: "Progress update",
		}
		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)

		mockReportProgressRepo.On("CreateTX", ctx, mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*model.ReportProgress")).
			Return(nil, errors.New("database error"))

		result, err := service.UploadProgressReport(ctx, db, 1, 1, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
		mockReportProgressRepo.AssertExpectations(t)
	})
}

func TestReportService_GetProgressReports(t *testing.T) {
	ctx := context.Background()

	t.Run("should get progress reports successfully", func(t *testing.T) {
		_, _, _, _, _, _, mockReportProgressRepo, _, _, _, service := setupMocks()

		progresses := []model.ReportProgress{
			{
				ID:        1,
				ReportID:  1,
				UserID:    1,
				Status:    "Progress 1",
				CreatedAt: time.Now().Unix(),
			},
		}

		mockReportProgressRepo.On("GetByReportID", ctx, uint(1)).Return(progresses, nil)

		result, err := service.GetProgressReports(ctx, 1)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 1)
		mockReportProgressRepo.AssertExpectations(t)
	})

	t.Run("should return not found error when no progress reports found", func(t *testing.T) {
		_, _, _, _, _, _, mockReportProgressRepo, _, _, _, service := setupMocks()

		mockReportProgressRepo.On("GetByReportID", ctx, uint(1)).Return(nil, gorm.ErrRecordNotFound)
		result, err := service.GetProgressReports(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "progres laporan tidak ditemukan", err.Error())
		mockReportProgressRepo.AssertExpectations(t)
	})

	t.Run("should return error when fetch fails", func(t *testing.T) {
		_, _, _, _, _, _, mockReportProgressRepo, _, _, _, service := setupMocks()

		mockReportProgressRepo.On("GetByReportID", ctx, uint(1)).Return(nil, errors.New("database error"))

		result, err := service.GetProgressReports(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportProgressRepo.AssertExpectations(t)
	})
}

func TestReportService_GetReportStatistics(t *testing.T) {
	ctx := context.Background()

	t.Run("should get report statistics successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		totalReportCount := &dto.TotalReportCount{
			TotalReports:               100,
			TotalInfrastructureReports: 20,
			TotalEnvironmentReports:    15,
			TotalSafetyReports:         10,
			TotalTrafficReports:        8,
			TotalPublicFacilityReports: 7,
			TotalWasteReports:          6,
			TotalWaterReports:          5,
			TotalElectricityReports:    4,
			TotalHealthReports:         3,
			TotalSocialReports:         2,
			TotalEducationReports:      1,
			TotalAdministrativeReports: 1,
			TotalDisasterReports:       1,
			TotalOtherReports:          17,
		}

		monthlyCount := map[string]int64{
			"2026-01": 10,
			"2026-02": 15,
		}

		statusCount := map[string]int64{
			"WAITING":              30,
			"ON_PROGRESS":          20,
			"NOT_RESOLVED":         15,
			"POTENTIALLY_RESOLVED": 10,
			"RESOLVED":             20,
			"EXPIRED":              5,
		}

		mockReportRepo.On("GetByReportTypeCount", ctx).Return(totalReportCount, nil)

		mockReportRepo.On("GetByReportStatusCount", ctx, []string{"WAITING", "ON_PROGRESS", "NOT_RESOLVED", "POTENTIALLY_RESOLVED", "RESOLVED", "EXPIRED"}).Return(statusCount, nil)

		mockReportRepo.On("GetMonthlyReportCount", ctx).Return(monthlyCount, nil)

		result, err := service.GetReportStatistics(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(100), result.TotalReports)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should return error when fetch reports count fails", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		mockReportRepo.On("GetByReportTypeCount", ctx).Return(nil, errors.New("database error"))

		result, err := service.GetReportStatistics(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_CreateReportComment(t *testing.T) {
	ctx := context.Background()

	t.Run("should create root comment successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, mockUserRepo, mockUserProfileRepo, _, _, _, mockReportCommentRepo, service := setupMocks()

		existingReport := &model.Report{
			ID:     1,
			UserID: 2,
		}

		user := &model.User{
			ID:       1,
			Username: "testuser",
		}

		userProfile := &model.UserProfile{
			UserID: 1,
		}

		req := dto.CreateReportCommentRequest{
			Content:         mainutils.StrPtrOrNil("Test comment"),
			ParentCommentID: nil,
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockUserRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
		mockUserProfileRepo.On("GetByUserID", ctx, uint(1)).Return(userProfile, nil)
		mockReportCommentRepo.On("Create", ctx, mock.AnythingOfType("*model.ReportComment")).
			Return(&model.ReportComment{ID: primitive.NewObjectID()}, nil)

		result, err := service.CreateReportComment(ctx, nil, 1, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test comment", *result.Content)
		mockReportRepo.AssertExpectations(t)
		mockReportCommentRepo.AssertExpectations(t)
	})

	t.Run("should create reply comment successfully", func(t *testing.T) {
		mockReportRepo, _, _, _, mockUserRepo, mockUserProfileRepo, _, _, _, mockReportCommentRepo, service := setupMocks()

		existingReport := &model.Report{
			ID:     1,
			UserID: 2,
		}

		user := &model.User{
			ID:       1,
			Username: "testuser",
		}

		userProfile := &model.UserProfile{
			UserID: 1,
		}
		parentCommentID := primitive.NewObjectID()

		req := dto.CreateReportCommentRequest{
			Content:         mainutils.StrPtrOrNil("Reply comment"),
			ParentCommentID: mainutils.StrPtrOrNil(parentCommentID.Hex()),
		}
		mockReportRepo.On("GetByID", ctx, uint(1)).Return(existingReport, nil)
		mockUserRepo.On("GetByID", ctx, uint(1)).Return(user, nil)
		mockUserProfileRepo.On("GetByUserID", ctx, uint(1)).Return(userProfile, nil)
		mockReportCommentRepo.On("Create", ctx, mock.AnythingOfType("*model.ReportComment")).
			Return(&model.ReportComment{ID: primitive.NewObjectID()}, nil)
		result, err := service.CreateReportComment(ctx, nil, 1, 1, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Reply comment", *result.Content)
		mockReportRepo.AssertExpectations(t)
		mockReportCommentRepo.AssertExpectations(t)
	})

	t.Run("should return error when report not found", func(t *testing.T) {
		mockReportRepo, _, _, _, _, _, _, _, _, _, service := setupMocks()

		req := dto.CreateReportCommentRequest{
			Content: mainutils.StrPtrOrNil("Test comment"),
		}

		mockReportRepo.On("GetByID", ctx, uint(1)).Return(nil, gorm.ErrRecordNotFound)

		result, err := service.CreateReportComment(ctx, nil, 1, 1, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestReportService_GetReportComments(t *testing.T) {
	ctx := context.Background()

	t.Run("should get root comments successfully", func(t *testing.T) {
		_, _, _, _, mockUserRepo, _, _, _, _, mockReportCommentRepo, service := setupMocks()

		comments := []*model.ReportComment{
			{
				ID:        primitive.NewObjectID(),
				UserID:    1,
				ReportID:  1,
				Content:   mainutils.StrPtrOrNil("Test comment"),
				CreatedAt: time.Now().Unix(),
			},
		}

		userIds := []uint{1}

		mockReportCommentRepo.On("GetPaginatedRootByReportID", ctx, uint(1), (*primitive.ObjectID)(nil), 51).
			Return(comments, nil)

		mockUserRepo.On("GetByIDs", ctx, userIds).Return([]model.User{
			{
				ID:       1,
				Username: "testuser",
			},
		}, nil)

		mockReportCommentRepo.On("GetCountsByRootID", ctx, mock.AnythingOfType("primitive.ObjectID")).Return(int64(0), nil)

		mockReportCommentRepo.On("GetCountsByReportID", ctx, uint(1)).Return(int64(1), nil)

		result, err := service.GetReportComments(ctx, 1, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Comments, 1)
		mockReportCommentRepo.AssertExpectations(t)
	})

	t.Run("should return error when fetching comments fails", func(t *testing.T) {
		_, _, _, _, _, _, _, _, _, mockReportCommentRepo, service := setupMocks()

		mockReportCommentRepo.On("GetPaginatedRootByReportID", ctx, uint(1), (*primitive.ObjectID)(nil), 51).
			Return(nil, errors.New("database error"))

		result, err := service.GetReportComments(ctx, 1, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportCommentRepo.AssertExpectations(t)
	})
}

func TestReportService_GetReportCommentReplies(t *testing.T) {
	ctx := context.Background()

	t.Run("should get comment replies successfully", func(t *testing.T) {
		_, _, _, _, mockUserRepo, _, _, _, _, mockReportCommentRepo, service := setupMocks()

		rootID := primitive.NewObjectID()
		replies := []*model.ReportComment{
			{
				ID:           primitive.NewObjectID(),
				UserID:       1,
				ReportID:     1,
				ThreadRootID: &rootID,
				Content:      mainutils.StrPtrOrNil("Reply comment"),
				CreatedAt:    time.Now().Unix(),
			},
		}

		userIds := []uint{1}

		mockReportCommentRepo.On("GetPaginatedRepliesByRootID", ctx, rootID, (*primitive.ObjectID)(nil), 61).
			Return(replies, nil)

		mockReportCommentRepo.On("GetByID", ctx, rootID).
			Return(&model.ReportComment{ID: rootID, UserID: 1}, nil)

		mockUserRepo.On("GetByIDs", ctx, userIds).Return([]model.User{
			{
				ID:       1,
				Username: "testuser",
			},
		}, nil)

		mockReportCommentRepo.On("GetCountsByRootID", ctx, rootID).Return(int64(1), nil)

		result, err := service.GetReportCommentReplies(ctx, rootID.Hex(), nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Replies, 1)
		mockReportCommentRepo.AssertExpectations(t)
	})

	t.Run("should return error when fetching replies fails", func(t *testing.T) {
		_, _, _, _, _, _, _, _, _, mockReportCommentRepo, service := setupMocks()
		rootID := primitive.NewObjectID()

		mockReportCommentRepo.On("GetPaginatedRepliesByRootID", ctx, rootID, (*primitive.ObjectID)(nil), 61).
			Return(nil, errors.New("database error"))
		result, err := service.GetReportCommentReplies(ctx, rootID.Hex(), nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockReportCommentRepo.AssertExpectations(t)
	})

	t.Run("should return error when root ID is invalid", func(t *testing.T) {
		_, _, _, _, _, _, _, _, _, _, service := setupMocks()

		result, err := service.GetReportCommentReplies(ctx, "invalid-id", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
