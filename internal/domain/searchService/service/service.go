package service

import (
	"context"
	"errors"
	reportRepository "pingspot/internal/domain/reportService/repository"
	"pingspot/internal/domain/searchService/dto"
	userRepository "pingspot/internal/domain/userService/repository"
	apperror "pingspot/pkg/apperror"
	"pingspot/pkg/logger"
	contextutils "pingspot/pkg/utils/contextUtils"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SearchService struct {
	userRepo   userRepository.UserRepository
	reportRepo reportRepository.ReportRepository
}

func NewSearchService(
	userRepo userRepository.UserRepository,
	reportRepo reportRepository.ReportRepository,
) *SearchService {
	return &SearchService{
		userRepo:   userRepo,
		reportRepo: reportRepo,
	}
}

func (s *SearchService) SearchData(ctx context.Context, searchQuery string, usersDataNextCursor, reportsDataNextCursor uint, limit int) (*dto.SearchResponse, error) {
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
		return nil, apperror.New(500, "USER_SEARCH_FAILED", "Gagal mencari data pengguna", err.Error())
	}

	reportsData, err := s.reportRepo.FullTextSearchReportPaginated(ctx, strings.ToLower(searchQuery), limit, reportsDataNextCursor)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Error("Failed to search reports",
			zap.String("request_id", requestID),
			zap.String("search_query", searchQuery),
			zap.Error(err),
		)
		return nil, apperror.New(500, "REPORT_SEARCH_FAILED", "Gagal mencari data laporan", err.Error())
	}

	resultUsers := make([]dto.UsersSearch, 0, len(*usersData))
	for _, user := range *usersData {
		userDTO := dto.UsersSearch{
			UserID:         user.ID,
			FullName:       user.FullName,
			Email: 			user.Email,
			Bio:            user.Profile.Bio,
			ProfilePicture: user.Profile.ProfilePicture,
			Username:	   user.Username,
			Birthday:   	   user.Profile.Birthday,
			Gender: user.Profile.Gender,	
		}
		resultUsers = append(resultUsers, userDTO)
	}

	resultReports := make([]dto.ReportsSearch, 0, len(*reportsData))
	for _, report := range *reportsData {
		reportDTO := dto.ReportsSearch{
			ID:                 report.ID,
			ReportTitle:       report.ReportTitle,
			ReportType:        string(report.ReportType),
			ReportDescription: report.ReportDescription,
			ReportHasProgress: *report.HasProgress,
			ReportStatus:      string(report.ReportStatus),
			CreatedAt:         report.CreatedAt,
			UpdatedAt:         report.UpdatedAt,
		}
		resultReports = append(resultReports, reportDTO)
	}

	searchResponse := dto.SearchResponse{
		UsersData:   dto.UserSearchResult{Users: resultUsers, Type: "users"},
		ReportsData: dto.ReportSearchResult{Reports: resultReports, Type: "reports"},
	}

	logger.Info("Search completed successfully",
		zap.String("request_id", requestID),
		zap.Int("users_found", len(*usersData)),
		zap.Int("reports_found", len(*reportsData)),
	)

	return &searchResponse, nil
}
