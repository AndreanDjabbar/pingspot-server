package handler

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"pingspot/internal/domain/report_service/dto"
	"pingspot/internal/domain/report_service/service"
	"pingspot/internal/domain/report_service/validation"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/main_util"
	response "pingspot/pkg/utils/response_util"
	tokenutils "pingspot/pkg/utils/token_util"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type ReportHandler struct {
	reportService *service.ReportService
}

func NewReportHandler(reportService *service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

func (h *ReportHandler) CreateReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}

	reportTitle := c.FormValue("reportTitle")
	reportDescription := c.FormValue("reportDescription")
	reportType := c.FormValue("reportType")
	detailLocation := c.FormValue("detailLocation")
	hasProgressStr := c.FormValue("hasProgress")
	latitude := c.FormValue("latitude")
	longitude := c.FormValue("longitude")
	displayName := c.FormValue("displayName")
	addressType := c.FormValue("addressType")
	road := c.FormValue("road")
	state := c.FormValue("state")
	country := c.FormValue("country")
	postCode := c.FormValue("postCode")
	mapZoom := c.FormValue("mapZoom")
	region := c.FormValue("region")
	countryCode := c.FormValue("countryCode")
	county := c.FormValue("county")
	village := c.FormValue("village")
	suburb := c.FormValue("suburb")
	totalImageSize := int64(0)
	var images map[int]string = make(map[int]string)

	mapZoomInt, err := mainutils.StringToInt(mapZoom)
	if err != nil && mapZoom != "" {
		logger.Error("Invalid mapZoom format", zap.String("mapZoom", mapZoom), zap.Error(err))
	}

	files := form.File["reportImages"]
	if len(files) > 5 {
		logger.Error("Too many report images", zap.Int("count", len(files)))
		return response.ResponseError(c, 400, "Too many images", "", "Maximum 5 images")
	}
	const maxFileSize = 2 * 1024 * 1024
	const maxTotalSize = 10 * 1024 * 1024

	for _, file := range files {
		if file.Size > maxFileSize {
			logger.Error("Report image file size too large",
				zap.Int64("size", file.Size),
			)
			return response.ResponseError(
				c,
				400,
				"One of the images is too large",
				"",
				fmt.Sprintf("Maximum image size is %dMB per image", maxFileSize/(1024*1024)),
			)
		}
		totalImageSize += file.Size
	}

	if totalImageSize > maxTotalSize {
		logger.Error("Total report images size too large",
			zap.Int64("total_size", totalImageSize),
		)
		return response.ResponseError(
			c,
			400,
			"Total image size is too large",
			"",
			fmt.Sprintf("Maximum total image size is %dMB", maxTotalSize/(1024*1024)),
		)
	}

	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	validMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	for i, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !validExtensions[ext] {
			logger.Error("Unsupported image extension", zap.String("extension", ext))
			return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
		}

		contentType := file.Header.Get("Content-Type")
		if !validMimeTypes[contentType] {
			logger.Error("Invalid content type", zap.String("mime", contentType))
			return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
		}

		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		images[i] = fileName
		savePath := filepath.Join("uploads/main/report", fileName)
		if err := c.SaveFile(file, savePath); err != nil {
			for j := range i {
				os.Remove(filepath.Join("uploads/main/report", images[j]))
			}
			logger.Error("Failed to save image", zap.Error(err))
			return response.ResponseError(c, 500, "Failed to save image", "", err.Error())
		}
	}

	floatLatitude, err := mainutils.StringToFloat64(latitude)
	if err != nil {
		logger.Error("Invalid latitude format", zap.String("latitude", latitude), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid latitude format", "", "Latitude must be a decimal number")
	}

	floatLongitude, err := mainutils.StringToFloat64(longitude)
	if err != nil {
		logger.Error("Invalid longitude format", zap.String("longitude", longitude), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid longitude format", "", "Longitude must be a decimal number")
	}

	hasProgress, err := mainutils.StringToBool(hasProgressStr)
	if err != nil && hasProgressStr != "" {
		logger.Error("Invalid hasProgress format", zap.String("hasProgress", hasProgressStr), zap.Error(err))
	}

	req := dto.CreateReportRequest{
		ReportTitle:       reportTitle,
		ReportType:        reportType,
		ReportDescription: reportDescription,
		DetailLocation:    detailLocation,
		HasProgress:       hasProgress,
		Latitude:          floatLatitude,
		Longitude:         floatLongitude,
		DisplayName:       mainutils.StrPtrOrNil(displayName),
		AddressType:       mainutils.StrPtrOrNil(addressType),
		Country:           mainutils.StrPtrOrNil(country),
		CountryCode:       mainutils.StrPtrOrNil(countryCode),
		MapZoom:           &mapZoomInt,
		Region:            mainutils.StrPtrOrNil(region),
		PostCode:          mainutils.StrPtrOrNil(postCode),
		County:            mainutils.StrPtrOrNil(county),
		State:             mainutils.StrPtrOrNil(state),
		Road:              mainutils.StrPtrOrNil(road),
		Village:           mainutils.StrPtrOrNil(village),
		Suburb:            mainutils.StrPtrOrNil(suburb),
		Image1URL:         mainutils.StrPtrOrNil(images[0]),
		Image2URL:         mainutils.StrPtrOrNil(images[1]),
		Image3URL:         mainutils.StrPtrOrNil(images[2]),
		Image4URL:         mainutils.StrPtrOrNil(images[3]),
		Image5URL:         mainutils.StrPtrOrNil(images[4]),
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatCreateReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	result, err := h.reportService.CreateReport(ctx, userID, req)
	if err != nil {
		for i := range files {
			os.Remove(filepath.Join("uploads/main/report", images[i]))
		}
		logger.Error("Failed to create report", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to create report", "", err.Error())
	}

	logger.Info("Report created successfully", zap.Uint("report_id", result.Report.ID))
	return response.ResponseSuccess(c, 200, "Report created successfully", "data", result)
}

func (h *ReportHandler) EditReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}

	reportTitle := c.FormValue("reportTitle")
	reportDescription := c.FormValue("reportDescription")
	reportType := c.FormValue("reportType")
	detailLocation := c.FormValue("detailLocation")
	hasProgressStr := c.FormValue("hasProgress")
	latitude := c.FormValue("latitude")
	longitude := c.FormValue("longitude")
	displayName := c.FormValue("displayName")
	addressType := c.FormValue("addressType")
	road := c.FormValue("road")
	state := c.FormValue("state")
	country := c.FormValue("country")
	mapZoom := c.FormValue("mapZoom")
	postCode := c.FormValue("postCode")
	region := c.FormValue("region")
	countryCode := c.FormValue("countryCode")
	county := c.FormValue("county")
	village := c.FormValue("village")
	suburb := c.FormValue("suburb")
	existingImagesSTR := c.FormValue("existingImages")

	mapZoomInt, err := mainutils.StringToInt(mapZoom)
	if err != nil && mapZoom != "" {
		logger.Error("Invalid mapZoom format", zap.String("mapZoom", mapZoom), zap.Error(err))
	}

	var existingImages []string
	if existingImagesSTR != "" {
		if err := json.Unmarshal([]byte(existingImagesSTR), &existingImages); err != nil {
			logger.Error("Failed to unmarshal existingImages", zap.String("existingImages", existingImagesSTR), zap.Error(err))
			return response.ResponseError(c, 400, "Invalid existingImages format", "", "existingImages must be an array of strings")
		}
	}

	files := form.File["reportImages"]
	totalImageLen := len(files) + len(existingImages)

	var images map[int]string = make(map[int]string)
	newImages := make([]map[string]multipart.FileHeader, 0)

	if totalImageLen > 5 {
		logger.Error("Too many report images", zap.Int("count", totalImageLen))
		return response.ResponseError(c, 400, "Too many images", "", "Maximum 5 images")
	}

	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	validMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}

	const maxFileSize = 2 * 1024 * 1024

	if len(existingImages) > 0 {
		var existingIDX = 0
		for i, imgURL := range existingImages {
			images[i] = imgURL
			existingIDX = i
		}

		for i, file := range files {
			if file.Size > maxFileSize {
				logger.Error("Report image file size too large", zap.Int64("size", files[i].Size))
				return response.ResponseError(c, 400, "One of the images is too large", "", "Maximum image size is 5MB per image")
			}

			ext := strings.ToLower(filepath.Ext(file.Filename))
			if !validExtensions[ext] {
				logger.Error("Unsupported image extension", zap.String("extension", ext))
				return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
			}

			contentType := file.Header.Get("Content-Type")
			if !validMimeTypes[contentType] {
				logger.Error("Invalid content type", zap.String("mime", contentType))
				return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
			}

			fileName := fmt.Sprintf("%d%d%d%s", time.Now().UnixNano(), i, uintReportID, ext)
			newImages = append(newImages, map[string]multipart.FileHeader{fileName: *file})
			images[existingIDX+1+i] = fileName
		}
	} else {
		for i, file := range files {
			if file.Size > maxFileSize {
				logger.Error("Report image file size too large", zap.Int64("size", files[i].Size))
			}
			ext := filepath.Ext(file.Filename)
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
				logger.Error("Unsupported profile picture file format", zap.String("extension", ext))
				return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
			}
			fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			newImages = append(newImages, map[string]multipart.FileHeader{fileName: *file})
			images[i] = fileName
		}
	}

	floatLatitude, err := mainutils.StringToFloat64(latitude)
	if err != nil {
		logger.Error("Invalid latitude format", zap.String("latitude", latitude), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid latitude format", "", "Latitude must be a decimal number")
	}

	floatLongitude, err := mainutils.StringToFloat64(longitude)
	if err != nil {
		logger.Error("Invalid longitude format", zap.String("longitude", longitude), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid longitude format", "", "Longitude must be a decimal number")
	}

	hasProgress, err := mainutils.StringToBool(hasProgressStr)
	if err != nil && hasProgressStr != "" {
		logger.Error("Invalid hasProgress format", zap.String("hasProgress", hasProgressStr), zap.Error(err))
	}

	req := dto.EditReportRequest{
		ReportTitle:       reportTitle,
		ReportType:        reportType,
		ReportDescription: reportDescription,
		DetailLocation:    detailLocation,
		HasProgress:       hasProgress,
		MapZoom:           &mapZoomInt,
		Latitude:          floatLatitude,
		Longitude:         floatLongitude,
		DisplayName:       mainutils.StrPtrOrNil(displayName),
		AddressType:       mainutils.StrPtrOrNil(addressType),
		Country:           mainutils.StrPtrOrNil(country),
		CountryCode:       mainutils.StrPtrOrNil(countryCode),
		Region:            mainutils.StrPtrOrNil(region),
		PostCode:          mainutils.StrPtrOrNil(postCode),
		County:            mainutils.StrPtrOrNil(county),
		State:             mainutils.StrPtrOrNil(state),
		Road:              mainutils.StrPtrOrNil(road),
		Village:           mainutils.StrPtrOrNil(village),
		Suburb:            mainutils.StrPtrOrNil(suburb),
		Image1URL:         mainutils.StrPtrOrNil(images[0]),
		Image2URL:         mainutils.StrPtrOrNil(images[1]),
		Image3URL:         mainutils.StrPtrOrNil(images[2]),
		Image4URL:         mainutils.StrPtrOrNil(images[3]),
		Image5URL:         mainutils.StrPtrOrNil(images[4]),
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatEditReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	for _, file := range newImages {
		for k, v := range file {
			savePath := filepath.Join("uploads/main/report", k)
			if err := c.SaveFile(&v, savePath); err != nil {
				logger.Error("Failed to save image", zap.Error(err))
				return response.ResponseError(c, 500, "Failed to save image", "", err.Error())
			}
		}
	}

	result, err := h.reportService.EditReport(ctx, userID, uintReportID, req)
	if err != nil {
		for _, file := range newImages {
			for k := range file {
				os.Remove(filepath.Join("uploads/main/report", k))
			}
		}
		logger.Error("Failed to edit report", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to edit report", "", err.Error())
	}

	logger.Info("Report editted successfully", zap.Uint("report_id", uintReportID))

	return response.ResponseSuccess(c, 200, "Report edited successfully", "data", result)
}

func (h *ReportHandler) GetReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportID := c.Query("reportID")
	cursorID := c.Query("cursorID")
	reportOwnerID := c.Query("userID")
	distance := c.Query("distance")
	reportType := c.Query("reportType")
	status := c.Query("status")
	sortBy := c.Query("sortBy")
	hasProgress := c.Query("hasProgress")

	var formattedDistance dto.Distance

	if err := json.Unmarshal([]byte(distance), &formattedDistance); err != nil && distance != "" {
		logger.Error("Invalid distance format", zap.String("distance", distance), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid distance format", "", "Distance must be JSON with distance, lat, and lng fields")
	}

	cursorIDUint, err := mainutils.StringToUint(cursorID)
	if err != nil && cursorID != "" {
		logger.Error("Invalid afterID format", zap.String("afterID", cursorID), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid afterID format", "", "afterID must be a number")
	}
	reportOwnerIDUint, err := mainutils.StringToUint(reportOwnerID)
	if err != nil && reportOwnerID != "" {
		logger.Error("Invalid userID format", zap.String("userID", reportOwnerID), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid userID format", "", "userID must be a number")
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	if reportID == "" {
		reports, err := h.reportService.GetAllReport(ctx, userID, cursorIDUint, reportOwnerIDUint, reportType, status, sortBy, hasProgress, formattedDistance)
		if err != nil {
			logger.Error("Failed to get all reports", zap.Error(err))
			if appErr, ok := err.(*apperror.AppError); ok {
				return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
			}
			return response.ResponseError(c, 500, "Failed to retrieve reports", "", err.Error())
		}
		var nextCursor *uint = nil
		if len(reports.Reports) > 0 {
			lastReport := reports.Reports[len(reports.Reports)-1]
			nextCursor = &lastReport.ID
		}
		mappedData := fiber.Map{
			"reports":    reports,
			"nextCursor": nextCursor,
		}
		return response.ResponseSuccess(c, 200, "Get all reports success", "data", mappedData)
	} else {
		uintReportID, err := mainutils.StringToUint(reportID)
		if err != nil {
			logger.Error("Invalid reportID format", zap.String("reportID", reportID), zap.Error(err))
			return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
		}

		report, err := h.reportService.GetReportByID(ctx, userID, uintReportID)
		if err != nil {
			logger.Error("Failed to get report by ID", zap.Uint("reportID", uintReportID), zap.Error(err))
			if appErr, ok := err.(*apperror.AppError); ok {
				return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
			}
			return response.ResponseError(c, 500, "Failed to retrieve reports", "", err.Error())
		}
		mappedData := fiber.Map{
			"report": report,
		}
		return response.ResponseSuccess(c, 200, "Get report success", "data", mappedData)
	}
}

func (h *ReportHandler) ReactionReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}

	var req dto.ReactionReportRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatReactionReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))
	reaction, err := h.reportService.ReactToReport(ctx, userID, uintReportID, req.ReactionType)
	if err != nil {
		logger.Error("Failed to react to report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to react to report", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Report reaction successful", "", reaction)
}

func (h *ReportHandler) VoteReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	var req dto.VoteReportRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatVoteReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}
	vote, err := h.reportService.VoteToReport(ctx, userID, uintReportID, req.VoteType)
	if err != nil {
		logger.Error("Failed to vote to report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to vote on report", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Report vote successful", "", vote)
}

func (h *ReportHandler) UploadProgressReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}

	var images map[int]string = make(map[int]string)
	status := c.FormValue("progressStatus")
	notes := c.FormValue("progressNotes")

	files := form.File["progressAttachments"]
	if len(files) > 2 {
		logger.Error("Too many progress attachments", zap.Int("count", len(files)))
		return response.ResponseError(c, 400, "Too many attachments", "", "Maximum 2 attachments")
	}

	totalImageSize := int64(0)
	for i, file := range files {
		if file.Size > 5*1024*1024 {
			logger.Error("Report image file size too large", zap.Int64("size", files[i].Size))
			return response.ResponseError(c, 400, "One of the images is too large", "", "Maximum image size is 5MB per image")
		}
		totalImageSize += file.Size
	}

	if totalImageSize > 10*1024*1024 {
		logger.Error("Total progress attachments size too large", zap.Int64("total_size", totalImageSize))
		return response.ResponseError(c, 400, "Total attachment size is too large", "", "Maximum total attachment size is 10MB")
	}

	if len(files) > 0 {
		for i, file := range files {
			ext := filepath.Ext(file.Filename)
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".pdf" {
				logger.Error("Unsupported progress attachment file format", zap.String("extension", ext))
				return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG, PNG, or PDF")
			}
			fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			savePath := filepath.Join("uploads/main/report/progress", fileName)
			if err := c.SaveFile(file, savePath); err != nil {
				logger.Error("Failed to save progress attachment", zap.Error(err))
				return response.ResponseError(c, 500, "Failed to save attachment", "", err.Error())
			}
			images[i] = fileName
		}
	}

	req := dto.UploadProgressReportRequest{
		Status:      status,
		Notes:       notes,
		Attachment1: mainutils.StrPtrOrNil(images[0]),
		Attachment2: mainutils.StrPtrOrNil(images[1]),
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatUploadProgressReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	newProgress, err := h.reportService.UploadProgressReport(ctx, userID, uintReportID, req)
	if err != nil {
		logger.Error("Failed to upload progress report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to upload report progress", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Report progress uploaded successfully", "data", newProgress)
}

func (h *ReportHandler) GetProgressReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}

	progressList, err := h.reportService.GetProgressReports(ctx, uintReportID)
	if err != nil {
		logger.Error("Failed to get progress reports", zap.Uint("reportID", uintReportID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve report progress", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Get progress reports success", "data", progressList)
}

func (h *ReportHandler) DeleteReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))
	err = h.reportService.DeleteReport(ctx, userID, uintReportID, "soft")
	if err != nil {
		logger.Error("Failed to delete report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to delete report", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Report deleted successfully", "data", fiber.Map{
		"reportID": uintReportID,
	})
}

func (h *ReportHandler) CreateReportCommentHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}

	content := c.FormValue("content")
	mediaURL := c.FormValue("mediaURL")
	mediaType := c.FormValue("mediaType")
	mediaWidthStr := c.FormValue("mediaWidth")
	mediaHeightStr := c.FormValue("mediaHeight")
	parentCommentIDStr := c.FormValue("parentCommentID")
	threadRootIDStr := c.FormValue("threadRootID")
	mentionsSTR := c.FormValue("mentions")
	files := form.File["mediaFile"]

	if len(files) != 0 && len(files) > 0 && mediaType == "" {
		logger.Error("Media type is required when media file is provided")
		return response.ResponseError(c, 400, "mediaType is required when uploading a media file", "", "Set mediaType according to the uploaded media file type")
	}

	if len(files) > 1 {
		logger.Error("Too many media files", zap.Int("count", len(files)))
		return response.ResponseError(c, 400, "Too many media files", "", "Only one media file may be uploaded")
	}

	var mentions []uint
	var mediaWidth, mediaHeight *uint

	mediaWidthVal, err := mainutils.StringToUint(mediaWidthStr)
	if err != nil && mediaWidthStr != "" {
		logger.Error("Invalid mediaWidth format", zap.String("mediaWidth", mediaWidthStr), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid mediaWidth format", "", "mediaWidth must be a number")
	}

	if mediaWidthStr != "" {
		mediaWidth = &mediaWidthVal
	}

	mediaHeightVal, err := mainutils.StringToUint(mediaHeightStr)
	if err != nil && mediaHeightStr != "" {
		logger.Error("Invalid mediaHeight format", zap.String("mediaHeight", mediaHeightStr), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid mediaHeight format", "", "mediaHeight must be a number")
	}

	if mediaHeightStr != "" {
		mediaHeight = &mediaHeightVal
	}

	if mentionsSTR != "" {
		if err := json.Unmarshal([]byte(mentionsSTR), &mentions); err != nil {
			logger.Error("Failed to unmarshal mentions", zap.String("mentions", mentionsSTR), zap.Error(err))
			return response.ResponseError(c, 400, "Invalid mentions format", "", "mentions must be an array of numbers")
		}
	}

	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}

	validMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/webp": true,
	}
	imageName := ""
	for _, file := range files {
		if file.Size > 3*1024*1024 {
			logger.Error("Media file size too large", zap.Int64("size", file.Size))
			return response.ResponseError(c, 400, "Media file is too large", "", "Maximum media file size is 3MB")
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !validExtensions[ext] {
			logger.Error("Unsupported image extension", zap.String("extension", ext))
			return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
		}

		contentType := file.Header.Get("Content-Type")
		if !validMimeTypes[contentType] {
			logger.Error("Invalid content type", zap.String("mime", contentType))
			return response.ResponseError(c, 400, "Unsupported file format", "", "Use JPG or PNG")
		}
		fileName := fmt.Sprintf("%d%d%s", time.Now().UnixNano(), uintReportID, ext)
		imageName = fileName
		savePath := filepath.Join("uploads/main/report/comments", fileName)
		if err := c.SaveFile(file, savePath); err != nil {
			logger.Error("Failed to save media file", zap.Error(err))
			return response.ResponseError(c, 500, "Failed to save media file", "", err.Error())
		}
	}

	if mediaType == "IMAGE" && imageName != "" {
		mediaURL = imageName
	}

	req := dto.CreateReportCommentRequest{
		Content:         mainutils.StrPtrOrNil(content),
		MediaURL:        mainutils.StrPtrOrNil(mediaURL),
		MediaType:       mainutils.StrPtrOrNil(mediaType),
		MediaWidth:      mediaWidth,
		MediaHeight:     mediaHeight,
		ParentCommentID: mainutils.StrPtrOrNil(parentCommentIDStr),
		ThreadRootID:    mainutils.StrPtrOrNil(threadRootIDStr),
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatCreateReportCommentValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}
	newComment, err := h.reportService.CreateReportComment(ctx, userID, uintReportID, req)
	if err != nil {
		os.Remove(filepath.Join("uploads/main/report/comments", imageName))
		logger.Error("Failed to create report comment", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
	}
	return response.ResponseSuccess(c, 200, "Report comment created successfully", "data", newComment)
}

func (h *ReportHandler) GetReportCommentsHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}
	cursorID := c.Query("cursorID")

	comments, err := h.reportService.GetReportComments(ctx, uintReportID, mainutils.StrPtrOrNil(cursorID))
	if err != nil {
		logger.Error("Failed to get report comments", zap.Uint("reportID", uintReportID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve report comments", "", err.Error())
	}
	var nextCursor *string = nil
	if comments.HasMore && len(comments.Comments) > 0 {
		lastComment := comments.Comments[len(comments.Comments)-1]
		nextCursor = mainutils.StrPtrOrNil(lastComment.CommentID)
	}

	mappedData := fiber.Map{
		"comments":   comments,
		"nextCursor": nextCursor,
	}
	return response.ResponseSuccess(c, 200, "Report comments retrieved successfully", "data", mappedData)
}

func (h *ReportHandler) GetReportStatisticsHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()

	reportStatistics, err := h.reportService.GetReportStatistics(ctx)
	if err != nil {
		logger.Error("Failed to get report statistics", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
	}
	return response.ResponseSuccess(c, 200, "Report statistics retrieved successfully", "data", reportStatistics)
}

func (h *ReportHandler) GetReportCommentRepliesHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	commentIDParam := c.Params("commentID")

	cursorID := c.Query("cursorID")

	replies, err := h.reportService.GetReportCommentReplies(ctx, commentIDParam, mainutils.StrPtrOrNil(cursorID))
	if err != nil {
		logger.Error("Failed to get report comment replies", zap.String("commentID", commentIDParam), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve report comment replies", "", err.Error())
	}
	var nextCursor *string = nil

	if replies.HasMore && len(replies.Replies) > 0 {
		lastComment := replies.Replies[len(replies.Replies)-1]
		nextCursor = mainutils.StrPtrOrNil(lastComment.CommentID)
	}
	mappedData := fiber.Map{
		"replies":    replies,
		"nextCursor": nextCursor,
	}
	return response.ResponseSuccess(c, 200, "Report comment replies retrieved successfully", "data", mappedData)
}

func (h *ReportHandler) SaveReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid reportID format", "", "reportID must be a number")
	}

	var req dto.SaveReportRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatSaveReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))
	save, err := h.reportService.SaveReport(ctx, userID, uintReportID, *req.Save)
	if err != nil {
		logger.Error("Failed to save report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to save report", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Report saved successfully", "", save)
}

func (h *ReportHandler) GetSavedReportsHandler(c *fiber.Ctx) error {
	cursorID := c.Query("cursorID")

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}
	userID := uint(claims["user_id"].(float64))

	savedReports, err := h.reportService.GetSavedReports(c.UserContext(), userID, mainutils.StrPtrOrNil(cursorID))
	if err != nil {
		logger.Error("Failed to get saved reports", zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve saved reports", "", err.Error())
	}

	var nextCursor *uint = nil
	if len(*savedReports.SavedReports) > 0 {
		lastReport := (*savedReports.SavedReports)[len(*savedReports.SavedReports)-1]
		nextCursor = &lastReport.ReportSavedID
	}

	mappedData := fiber.Map{
		"savedReports": savedReports,
		"nextCursor":   nextCursor,
	}
	return response.ResponseSuccess(c, 200, "Saved reports retrieved successfully", "data", mappedData)
}
