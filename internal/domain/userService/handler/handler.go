package handler

import (
	"fmt"
	"path/filepath"
	"pingspot/internal/domain/userService/dto"
	"pingspot/internal/domain/userService/service"
	"pingspot/internal/domain/userService/validation"
	"pingspot/internal/infrastructure/database"
	apperror "pingspot/pkg/apperror"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/mainUtils"
	"pingspot/pkg/utils/response"
	tokenutils "pingspot/pkg/utils/tokenutils"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) SaveUserSecurityHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dto.SaveUserSecurityRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatSaveUserSecurityValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userId := uint(claims["user_id"].(float64))
	if err := h.userService.SaveSecurity(ctx, userId, req); err != nil {
		logger.Error("Failed to update user password", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memperbarui kata sandi", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Kata sandi berhasil diperbarui. Silahkan masuk kembali dengan kata sandi baru anda.", "data", nil)
}

func (h *UserHandler) SaveUserProfileHandler(c *fiber.Ctx) error {
	_, err := c.MultipartForm()
	if err != nil {
		logger.Error("Failed to parse multipart form", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}
	fullName := c.FormValue("fullName")
	gender := c.FormValue("gender")
	birthday := c.FormValue("birthday")
	bio := c.FormValue("bio")
	username := c.FormValue("username")
	file, err := c.FormFile("profilePicture")
	var profilePicture string
	if err == nil && file != nil {
		if file.Size > 5*1024*1024 {
			logger.Error("Profile picture file size too large", zap.Int64("size", file.Size))
			return response.ResponseError(c, 400, "Ukuran gambar terlalu besar", "", "Maksimal ukuran gambar 5MB")
		}

		ext := filepath.Ext(file.Filename)
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			logger.Error("Unsupported profile picture file format", zap.String("extension", ext))
			return response.ResponseError(c, 400, "Format file tidak didukung", "", "Gunakan JPG atau PNG")
		}

		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
		savePath := filepath.Join("uploads/user", fileName)

		if err := c.SaveFile(file, savePath); err != nil {
			logger.Error("Failed to save profile picture", zap.Error(err))
			return response.ResponseError(c, 500, "Gagal menyimpan gambar", "", err.Error())
		}
		profilePicture = fileName
	} else {
		if c.FormValue("removeProfilePicture") == "true" {
			profilePicture = ""
		} else if c.FormValue("defaultProfilePicture") != "" {
			profilePictureName := c.FormValue("defaultProfilePicture")
			profilePicture = profilePictureName
		} else {
			profilePicture = ""
		}
	}

	req := dto.SaveUserProfileRequest{
		FullName:       fullName,
		Gender:         mainutils.StrPtrOrNil(gender),
		Bio:            mainutils.StrPtrOrNil(bio),
		ProfilePicture: mainutils.StrPtrOrNil(profilePicture),
		Birthday:       mainutils.StrPtrOrNil(birthday),
		Username:       mainutils.StrPtrOrNil(username),
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatSaveUserProfileValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userId := uint(claims["user_id"].(float64))
	ctx := c.UserContext()
	database := database.GetPostgresDB()
	newProfile, err := h.userService.SaveProfile(ctx, database, userId, req)
	if err != nil {
		logger.Error("Failed to save user profile", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memperbarui profil pengguna", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Profil pengguna berhasil diperbarui", "data", newProfile)
}

func (h *UserHandler) GetProfileHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}
	userId := uint(claims["user_id"].(float64))
	userProfile, err := h.userService.GetProfile(ctx, userId)
	if err != nil {
		logger.Error("Failed to get my profile", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan profil pengguna", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Berhasil mendapatkan profil pengguna", "data", userProfile)
}

func (h *UserHandler) GetUserStatistics(c *fiber.Ctx) error {
	ctx := c.UserContext()
	userStatistics, err := h.userService.GetUserStatistics(ctx)
	if err != nil {
		logger.Error("Failed to get user statistics", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan statistik pengguna", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Berhasil mendapatkan statistik pengguna", "data", userStatistics)
}

func (h *UserHandler) GetProfileByUsernameHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	username := c.Params("username")
	userProfile, err := h.userService.GetProfileByUsername(ctx, username)
	if err != nil {
		logger.Error("Failed to get user profile by username", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan profil pengguna", "", err.Error())
	}
	return response.ResponseSuccess(c, 200, "Berhasil mendapatkan profil pengguna", "data", userProfile)
}