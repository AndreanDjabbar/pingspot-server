package service

import (
	"context"
	"pingspot/internal/mocks/report"
	userMocks "pingspot/internal/mocks/user"
	"pingspot/internal/model"
	mainutils "pingspot/pkg/utils/mainUtils"
	"testing"

	"github.com/stretchr/testify/require"
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
	)
	require.NoError(t, err)

	return db
}

func TestNewSearchService(t *testing.T) {
	t.Run("should create new report service", func(t *testing.T) {
		mockReportRepo := new(report.MockReportRepository)
		mockUserRepo := new(userMocks.MockUserRepository)
		service := NewSearchService(
			mockUserRepo,
			mockReportRepo,
		)
		require.NotNil(t, service)
	})
}

func setupMocks() (
	*userMocks.MockUserRepository,
	*report.MockReportRepository,
	*SearchService,
) {
	mockReportRepo := new(report.MockReportRepository)
	mockUserRepo := new(userMocks.MockUserRepository)

	service := NewSearchService(
		mockUserRepo,
		mockReportRepo,
	)

	return mockUserRepo, mockReportRepo, service
}

func TestSearchService_SearchData(t *testing.T) {
	ctx := context.Background()

	t.Run("should return search results successfully", func(t *testing.T) {
		mockUserRepo, mockReportRepo, service := setupMocks()

		searchQuery := "test"
		limit := 10
		var usersDataNextCursor uint = 0
		var reportsDataNextCursor uint = 0

		hasProgress := true

		mockUserRepo.On("FullTextSearchUsersPaginated", ctx, searchQuery, limit+1, usersDataNextCursor).
			Return(&[]model.User{
				{
					ID:       1,
					Username: "testuser1",
					FullName: "Test User 1",
					Email:    "test1@example.com",
					Profile: model.UserProfile{
						UserID: 1,
						Bio:    mainutils.StrPtrOrNil("test_bio1"),
						Gender: mainutils.StrPtrOrNil("male"),
					},
				},
				{
					ID:       2,
					Username: "testuser2",
					FullName: "Test User 2",
					Email:    "test2@example.com",
					Profile: model.UserProfile{
						UserID: 2,
						Bio:    mainutils.StrPtrOrNil("test_bio2"),
						Gender: mainutils.StrPtrOrNil("female"),
					},
				},
			}, nil)

		mockReportRepo.On("FullTextSearchReportPaginated", ctx, searchQuery, limit+1, reportsDataNextCursor).
			Return(&[]model.Report{
				{
					ID:                1,
					ReportTitle:       "Test Report 1",
					ReportDescription: "Description 1",
					ReportType:        model.Infrastructure,
					ReportStatus:      model.WAITING,
					HasProgress:       &hasProgress,
				},
				{
					ID:                2,
					ReportTitle:       "Test Report 2",
					ReportDescription: "Description 2",
					ReportType:        model.Environment,
					ReportStatus:      model.ON_PROGRESS,
					HasProgress:       &hasProgress,
				},
			}, nil)

		result, err := service.SearchData(ctx, searchQuery, usersDataNextCursor, reportsDataNextCursor, limit+1)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Len(t, result.UsersData.Users, 2)
		require.Len(t, result.ReportsData.Reports, 2)
		mockUserRepo.AssertExpectations(t)
		mockReportRepo.AssertExpectations(t)
	})

	t.Run("should handle error when searching users", func(t *testing.T) {
		mockUserRepo, mockReportRepo, service := setupMocks()
		searchQuery := "test"
		limit := 10
		var usersDataNextCursor uint = 0
		var reportsDataNextCursor uint = 0
		mockUserRepo.On("FullTextSearchUsersPaginated", ctx, searchQuery, limit+1, usersDataNextCursor).
			Return(nil, gorm.ErrInvalidData)
		_, err := service.SearchData(ctx, searchQuery, usersDataNextCursor, reportsDataNextCursor, limit+1)
		require.Error(t, err)
		mockUserRepo.AssertExpectations(t)
		mockReportRepo.AssertNotCalled(t, "FullTextSearchReportPaginated", ctx, searchQuery, limit+1, reportsDataNextCursor)
	})
}
