package handler

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"pingspot/internal/domain/reportService/dto"
	"pingspot/internal/domain/reportService/service"
	"pingspot/internal/domain/reportService/validation"
	"pingspot/internal/infrastructure/database"
	apperror "pingspot/pkg/apperror"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/mainUtils"
	"pingspot/pkg/utils/response"
	"pingspot/pkg/utils/tokenutils"
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
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
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
		return response.ResponseError(c, 400, "Terlalu banyak gambar", "", "Maksimal 5 gambar")
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
				"Ukuran salah satu gambar terlalu besar",
				"",
				fmt.Sprintf("Maksimal ukuran gambar %dMB per gambar", maxFileSize/(1024*1024)),
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
			"Total ukuran semua gambar terlalu besar",
			"",
			fmt.Sprintf("Maksimal total ukuran gambar %dMB", maxTotalSize/(1024*1024)),
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
			return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
		}

		contentType := file.Header.Get("Content-Type")
		if !validMimeTypes[contentType] {
			logger.Error("Invalid content type", zap.String("mime", contentType))
			return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
		}

		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		images[i] = fileName
		savePath := filepath.Join("uploads/main/report", fileName)
		if err := c.SaveFile(file, savePath); err != nil {
			for j := range i {
				os.Remove(filepath.Join("uploads/main/report", images[j]))
			}
			logger.Error("Failed to save image", zap.Error(err))
			return response.ResponseError(c, 500, "Gagal menyimpan gambar", "", err.Error())
		}
	}

	floatLatitude, err := mainutils.StringToFloat64(latitude)
	if err != nil {
		logger.Error("Invalid latitude format", zap.String("latitude", latitude), zap.Error(err))
		return response.ResponseError(c, 400, "Format latitude tidak valid", "", "Latitude harus berupa angka desimal")
	}

	floatLongitude, err := mainutils.StringToFloat64(longitude)
	if err != nil {
		logger.Error("Invalid longitude format", zap.String("longitude", longitude), zap.Error(err))
		return response.ResponseError(c, 400, "Format longitude tidak valid", "", "Longitude harus berupa angka desimal")
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

	db := database.GetPostgresDB()
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatCreateReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))

	result, err := h.reportService.CreateReport(ctx, db, userID, req)
	if err != nil {
		for i := range files {
			os.Remove(filepath.Join("uploads/main/report", images[i]))
		}
		logger.Error("Failed to create report", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal membuat laporan", "", err.Error())
	}

	logger.Info("Report created successfully", zap.Uint("report_id", result.Report.ID))
	return response.ResponseSuccess(c, 200, "Laporan berhasil dibuat", "data", result)
}

func (h *ReportHandler) EditReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
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
			return response.ResponseError(c, 400, "Format existingImages tidak valid", "", "existingImages harus berupa array of string")
		}
	}

	files := form.File["reportImages"]
	totalImageLen := len(files) + len(existingImages)

	var images map[int]string = make(map[int]string)
	newImages := make([]map[string]multipart.FileHeader, 0)

	if totalImageLen > 5 {
		logger.Error("Too many report images", zap.Int("count", totalImageLen))
		return response.ResponseError(c, 400, "Terlalu banyak gambar", "", "Maksimal 5 gambar")
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
				return response.ResponseError(c, 400, "Ukuran salah satu gambar terlalu besar", "", "Maksimal ukuran gambar 5MB per gambar")
			}

			ext := strings.ToLower(filepath.Ext(file.Filename))
			if !validExtensions[ext] {
				logger.Error("Unsupported image extension", zap.String("extension", ext))
				return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
			}

			contentType := file.Header.Get("Content-Type")
			if !validMimeTypes[contentType] {
				logger.Error("Invalid content type", zap.String("mime", contentType))
				return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
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
				return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
			}
			fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			newImages = append(newImages, map[string]multipart.FileHeader{fileName: *file})
			images[i] = fileName
		}
	}

	floatLatitude, err := mainutils.StringToFloat64(latitude)
	if err != nil {
		logger.Error("Invalid latitude format", zap.String("latitude", latitude), zap.Error(err))
		return response.ResponseError(c, 400, "Format latitude tidak valid", "", "Latitude harus berupa angka desimal")
	}

	floatLongitude, err := mainutils.StringToFloat64(longitude)
	if err != nil {
		logger.Error("Invalid longitude format", zap.String("longitude", longitude), zap.Error(err))
		return response.ResponseError(c, 400, "Format longitude tidak valid", "", "Longitude harus berupa angka desimal")
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

	db := database.GetPostgresDB()
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatEditReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))

	for _, file := range newImages {
		for k, v := range file {
			savePath := filepath.Join("uploads/main/report", k)
			if err := c.SaveFile(&v, savePath); err != nil {
				logger.Error("Failed to save image", zap.Error(err))
				return response.ResponseError(c, 500, "Gagal menyimpan gambar", "", err.Error())
			}
		}
	}

	result, err := h.reportService.EditReport(ctx, db, userID, uintReportID, req)
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
		return response.ResponseError(c, 500, "Gagal menyunting laporan", "", err.Error())
	}

	logger.Info("Report editted successfully", zap.Uint("report_id", uintReportID))

	return response.ResponseSuccess(c, 200, "Laporan berhasil disunting", "data", result)
}

func (h *ReportHandler) GetReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportID := c.Query("reportID")
	cursorID := c.Query("cursorID")
	distance := c.Query("distance")
	reportType := c.Query("reportType")
	status := c.Query("status")
	sortBy := c.Query("sortBy")
	hasProgress := c.Query("hasProgress")

	var formattedDistance dto.Distance

	if err := json.Unmarshal([]byte(distance), &formattedDistance); err != nil && distance != "" {
		logger.Error("Invalid distance format", zap.String("distance", distance), zap.Error(err))
		return response.ResponseError(c, 400, "Format distance tidak valid", "", "Distance harus berupa JSON dengan field distance, lat, dan lng")
	}

	cursorIDUint, err := mainutils.StringToUint(cursorID)
	if err != nil && cursorID != "" {
		logger.Error("Invalid afterID format", zap.String("afterID", cursorID), zap.Error(err))
		return response.ResponseError(c, 400, "Format afterID tidak valid", "", "afterID harus berupa angka")
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))

	if reportID == "" {
		reports, err := h.reportService.GetAllReport(ctx, userID, cursorIDUint, reportType, status, sortBy, hasProgress, formattedDistance)
		if err != nil {
			logger.Error("Failed to get all reports", zap.Error(err))
			if appErr, ok := err.(*apperror.AppError); ok {
				return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
			}
			return response.ResponseError(c, 500, "Gagal mendapatkan laporan", "", err.Error())
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
			return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
		}

		report, err := h.reportService.GetReportByID(ctx, userID, uintReportID)
		if err != nil {
			logger.Error("Failed to get report by ID", zap.Uint("reportID", uintReportID), zap.Error(err))
			if appErr, ok := err.(*apperror.AppError); ok {
				return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
			}
			return response.ResponseError(c, 500, "Gagal mendapatkan laporan", "", err.Error())
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
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}

	var req dto.ReactionReportRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatReactionReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))
	db := database.GetPostgresDB()

	reaction, err := h.reportService.ReactToReport(ctx, db, userID, uintReportID, req.ReactionType)
	if err != nil {
		logger.Error("Failed to react to report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mereaksi laporan", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Reaksi laporan berhasil", "", reaction)
}

func (h *ReportHandler) VoteReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}
	var req dto.VoteReportRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatVoteReportValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))
	db := database.GetPostgresDB()
	vote, err := h.reportService.VoteToReport(ctx, db, userID, uintReportID, req.VoteType)
	if err != nil {
		logger.Error("Failed to vote to report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal vote laporan", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Vote laporan berhasil", "", vote)
}

func (h *ReportHandler) UploadProgressReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}

	var images map[int]string = make(map[int]string)
	status := c.FormValue("progressStatus")
	notes := c.FormValue("progressNotes")

	files := form.File["progressAttachments"]
	if len(files) > 2 {
		logger.Error("Too many progress attachments", zap.Int("count", len(files)))
		return response.ResponseError(c, 400, "Terlalu banyak lampiran", "", "Maksimal 2 lampiran")
	}

	totalImageSize := int64(0)
	for i, file := range files {
		if file.Size > 5*1024*1024 {
			logger.Error("Report image file size too large", zap.Int64("size", files[i].Size))
			return response.ResponseError(c, 400, "Ukuran salah satu gambar terlalu besar", "", "Maksimal ukuran gambar 5MB per gambar")
		}
		totalImageSize += file.Size
	}

	if totalImageSize > 10*1024*1024 {
		logger.Error("Total progress attachments size too large", zap.Int64("total_size", totalImageSize))
		return response.ResponseError(c, 400, "Ukuran total lampiran terlalu besar", "", "Maksimal ukuran total lampiran 10MB")
	}

	if len(files) > 0 {
		for i, file := range files {
			ext := filepath.Ext(file.Filename)
			if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".pdf" {
				logger.Error("Unsupported progress attachment file format", zap.String("extension", ext))
				return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG, PNG, atau PDF")
			}
			fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
			savePath := filepath.Join("uploads/main/report/progress", fileName)
			if err := c.SaveFile(file, savePath); err != nil {
				logger.Error("Failed to save progress attachment", zap.Error(err))
				return response.ResponseError(c, 500, "Gagal menyimpan lampiran", "", err.Error())
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
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))

	db := database.GetPostgresDB()

	newProgress, err := h.reportService.UploadProgressReport(ctx, db, userID, uintReportID, req)
	if err != nil {
		logger.Error("Failed to upload progress report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mengunggah progres laporan", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Progres laporan berhasil diunggah", "data", newProgress)
}

func (h *ReportHandler) GetProgressReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}

	progressList, err := h.reportService.GetProgressReports(ctx, uintReportID)
	if err != nil {
		logger.Error("Failed to get progress reports", zap.Uint("reportID", uintReportID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan progres laporan", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Get progress reports success", "data", progressList)
}

func (h *ReportHandler) DeleteReportHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))
	db := database.GetPostgresDB()
	err = h.reportService.DeleteReport(ctx, db, userID, uintReportID, "soft")
	if err != nil {
		logger.Error("Failed to delete report", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal menghapus laporan", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Laporan berhasil dihapus", "data", fiber.Map{
		"reportID": uintReportID,
	})
}

func (h *ReportHandler) CreateReportCommentHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userID := uint(claims["user_id"].(float64))

	form, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
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
		return response.ResponseError(c, 400, "mediaType wajib diisi jika mengunggah file media", "", "Isi mediaType sesuai dengan jenis file media yang diunggah")
	}

	if len(files) > 1 {
		logger.Error("Too many media files", zap.Int("count", len(files)))
		return response.ResponseError(c, 400, "Terlalu banyak file media", "", "Hanya boleh mengunggah 1 file media")
	}

	var mentions []uint
	var mediaWidth, mediaHeight *uint

	mediaWidthVal, err := mainutils.StringToUint(mediaWidthStr)
	if err != nil && mediaWidthStr != "" {
		logger.Error("Invalid mediaWidth format", zap.String("mediaWidth", mediaWidthStr), zap.Error(err))
		return response.ResponseError(c, 400, "Format mediaWidth tidak valid", "", "mediaWidth harus berupa angka")
	}

	if mediaWidthStr != "" {
		mediaWidth = &mediaWidthVal
	}

	mediaHeightVal, err := mainutils.StringToUint(mediaHeightStr)
	if err != nil && mediaHeightStr != "" {
		logger.Error("Invalid mediaHeight format", zap.String("mediaHeight", mediaHeightStr), zap.Error(err))
		return response.ResponseError(c, 400, "Format mediaHeight tidak valid", "", "mediaHeight harus berupa angka")
	}

	if mediaHeightStr != "" {
		mediaHeight = &mediaHeightVal
	}

	if mentionsSTR != "" {
		if err := json.Unmarshal([]byte(mentionsSTR), &mentions); err != nil {
			logger.Error("Failed to unmarshal mentions", zap.String("mentions", mentionsSTR), zap.Error(err))
			return response.ResponseError(c, 400, "Format mentions tidak valid", "", "mentions harus berupa array of angka")
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
			return response.ResponseError(c, 400, "Ukuran file media terlalu besar", "", "Maksimal ukuran file media 3MB")
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !validExtensions[ext] {
			logger.Error("Unsupported image extension", zap.String("extension", ext))
			return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
		}

		contentType := file.Header.Get("Content-Type")
		if !validMimeTypes[contentType] {
			logger.Error("Invalid content type", zap.String("mime", contentType))
			return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
		}
		fileName := fmt.Sprintf("%d%d%s", time.Now().UnixNano(), uintReportID, ext)
		imageName = fileName
		savePath := filepath.Join("uploads/main/report/comments", fileName)
		if err := c.SaveFile(file, savePath); err != nil {
			logger.Error("Failed to save media file", zap.Error(err))
			return response.ResponseError(c, 500, "Gagal menyimpan file media", "", err.Error())
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
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}
	db := database.GetMongoDB()

	newComment, err := h.reportService.CreateReportComment(ctx, db, userID, uintReportID, req)
	if err != nil {
		os.Remove(filepath.Join("uploads/main/report/comments", imageName))
		logger.Error("Failed to create report comment", zap.Uint("reportID", uintReportID), zap.Uint("userID", userID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
	}
	return response.ResponseSuccess(c, 200, "Komentar laporan berhasil dibuat", "data", newComment)
}

func (h *ReportHandler) GetReportCommentsHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	reportIDParam := c.Params("reportID")
	uintReportID, err := mainutils.StringToUint(reportIDParam)
	if err != nil {
		logger.Error("Invalid reportID format", zap.String("reportID", reportIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format reportID tidak valid", "", "reportID harus berupa angka")
	}
	cursorID := c.Query("cursorID")

	comments, err := h.reportService.GetReportComments(ctx, uintReportID, mainutils.StrPtrOrNil(cursorID))
	if err != nil {
		logger.Error("Failed to get report comments", zap.Uint("reportID", uintReportID), zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan komentar laporan", "", err.Error())
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
	return response.ResponseSuccess(c, 200, "Berhasil mengambil komentar laporan", "data", mappedData)
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
	return response.ResponseSuccess(c, 200, "Berhasil mengambil statistik laporan", "data", reportStatistics)
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
		return response.ResponseError(c, 500, "Gagal mendapatkan balasan komentar laporan", "", err.Error())
	}
	var nextCursor *string = nil
	
	if replies.HasMore && len(replies.Replies) > 0 {
		lastComment := replies.Replies[len(replies.Replies)-1]
		nextCursor = mainutils.StrPtrOrNil(lastComment.CommentID)
	}
	mappedData := fiber.Map{
		"replies":   replies,
		"nextCursor": nextCursor,
	}
	return response.ResponseSuccess(c, 200, "Sukses mengambil balasan komentar laporan", "data", mappedData)
}
