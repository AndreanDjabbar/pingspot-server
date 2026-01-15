package handler

import (
	"pingspot/internal/domain/searchService/service"
	"pingspot/pkg/logger"
	contextutils "pingspot/pkg/utils/contextUtils"
	mainutils "pingspot/pkg/utils/mainUtils"
	"pingspot/pkg/utils/response"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) HandleSearch(c *fiber.Ctx) error {
	const defaultLimit = 10
	ctx := c.UserContext()
	requestID := contextutils.GetRequestID(ctx)

	searchQuery := c.Query("searchQuery", "")
	if len(searchQuery) < 3 {
		logger.Warn("Search query too short",
			zap.String("request_id", requestID),
			zap.String("search_query", searchQuery),
		)
		return response.ResponseError(c, 400, "Panjang search query minimal 3 karakter", "", nil)
	}

	usersDataCursorID := c.Query("usersDataCursorID", "")
	usersDatacursorIDUint, err := mainutils.StringToUint(usersDataCursorID)
	if err != nil && usersDataCursorID != "" {
		logger.Error("Invalid afterID format", zap.String("afterID", usersDataCursorID), zap.Error(err))
		return response.ResponseError(c, 400, "Format afterID tidak valid", "", "afterID harus berupa angka")
	}

	reportsDataCursorID := c.Query("reportsDataCursorID", "")
	reportsDatacursorIDUint, err := mainutils.StringToUint(reportsDataCursorID)
	if err != nil && reportsDataCursorID != "" {
		logger.Error("Invalid afterID format", zap.String("afterID", reportsDataCursorID), zap.Error(err))
		return response.ResponseError(c, 400, "Format afterID tidak valid", "", "afterID harus berupa angka")
	}

	searchData, err := h.searchService.SearchData(ctx, searchQuery, usersDatacursorIDUint, reportsDatacursorIDUint, defaultLimit)
	if err != nil {
		logger.Error("Search failed",
			zap.String("request_id", requestID),
			zap.String("search_query", searchQuery),
			zap.Error(err),
		)
		return response.ResponseError(c, 500, "Gagal melakukan pencarian", err.Error(), nil)
	}

	logger.Info("Search request completed successfully",
		zap.String("request_id", requestID),
		zap.String("search_query", searchQuery),
	)

	var nextCursorUsersData *uint = nil
	if len(searchData.UsersData.Users) > 0 {
		lastUser := searchData.UsersData.Users[len(searchData.UsersData.Users)-1]
		nextCursorUsersData = &lastUser.UserID
	} 

	var nextCursorReportsData *uint = nil
	if len(searchData.ReportsData.Reports) > 0 {
		lastReport := searchData.ReportsData.Reports[len(searchData.ReportsData.Reports)-1]
		nextCursorReportsData = &lastReport.ID
	}

	finalResults := fiber.Map{
		"usersData":   searchData.UsersData,
		"nextCursorUsersData":   nextCursorUsersData,
		"reportsData": searchData.ReportsData,
		"nextCursorReportsData": nextCursorReportsData,
	}

	return response.ResponseSuccess(c, 200, "Pencarian berhasil", "data", finalResults)
}
