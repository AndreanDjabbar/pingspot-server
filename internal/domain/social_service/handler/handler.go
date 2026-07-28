package handler

import (
	"pingspot/internal/domain/social_service/service"
	"pingspot/internal/domain/user_service/dto"
	"pingspot/internal/domain/user_service/validation"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/main_util"
	response "pingspot/pkg/utils/response_util"
	tokenutils "pingspot/pkg/utils/token_util"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type SocialHandler struct {
	socialService *service.SocialService
}

func NewSocialHandler(socialService *service.SocialService) *SocialHandler {
	return &SocialHandler{socialService: socialService}
}

func (h *SocialHandler) FollowHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Token tidak valid", "", "Anda harus login terlebih dahulu")
	}

	userId := uint(claims["user_id"].(float64))

	var req dto.FollowRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatFollowValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	followingResult, err := h.socialService.Follow(ctx, userId, req)
	if err != nil {
		logger.Error("Failed to follow user", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mengikuti pengguna", "", err.Error())
	}
	var successMessage string = "Berhasil mengikuti pengguna"
	if followingResult.FollowProcess == "unfollow" {
		successMessage = "Berhasil berhenti mengikuti pengguna"
	}
	return response.ResponseSuccess(c, 200, successMessage, "data", followingResult)
}

func (h *SocialHandler) GetFollowDataHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()

	followingIDParam := c.Params("followingID")
	followingType := c.Params("followingType")

	followingID, err := mainutils.StringToUint(followingIDParam)
	if err != nil {
		logger.Error("Invalid followingID format", zap.String("followingID", followingIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Format followingID tidak valid", "", "followingID harus berupa angka")
	}

	var req dto.GetFollowDataRequest
	req.FollowingID = followingID
	req.FollowingType = followingType
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatGetFollowDataValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	followingData, err := h.socialService.GetFollowing(ctx, followingID, followingType)
	if err != nil {
		logger.Error("Failed to get following data", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan data following", "", err.Error())
	}

	return response.ResponseSuccess(c, 200, "Berhasil mendapatkan data following", "data", followingData)
}