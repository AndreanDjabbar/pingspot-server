package handler

import (
	"pingspot/internal/domain/social_service/dto"
	"pingspot/internal/domain/social_service/service"
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
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}

	userId := uint(claims["user_id"].(float64))

	var req dto.FollowRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Invalid request body format", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatFollowValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	followingResult, err := h.socialService.Follow(ctx, userId, req)
	if err != nil {
		logger.Error("Failed to follow user", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to follow user", "", err.Error())
	}
	var successMessage string = "Followed user successfully"
	if followingResult.FollowProcess == "unfollow" {
		successMessage = "Unfollowed user successfully"
	}
	return response.ResponseSuccess(c, 200, successMessage, "data", followingResult)
}

func (h *SocialHandler) GetFollowDataHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()

	claims, err := tokenutils.GetJWTClaims(c)
	if err != nil {
		logger.Error("Failed to get JWT claims", zap.Error(err))
		return response.ResponseError(c, 401, "Invalid token", "", "You must log in first")
	}

	userID := uint(claims["user_id"].(float64))

	followingIDParam := c.Params("followingID")
	followingType := c.Params("followingType")

	followingID, err := mainutils.StringToUint(followingIDParam)
	if err != nil {
		logger.Error("Invalid followingID format", zap.String("followingID", followingIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid followingID format", "", "followingID must be a number")
	}

	var req dto.GetFollowDataRequest
	req.FollowingID = followingID
	req.FollowingType = followingType
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatGetFollowDataValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validation failed", "errors", errors)
	}

	followingData, err := h.socialService.GetFollowing(ctx, followingID, followingType, userID)
	if err != nil {
		logger.Error("Failed to get following data", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve following data", "", err.Error())
	}

	return response.ResponseSuccess(c, 200, "Following data retrieved successfully", "data", followingData)
}

func (h *SocialHandler) GetUserConnectionsFollowersHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userIDParam := c.Params("userID")
	cursorID := c.Query("cursorID")

	userID, err := mainutils.StringToUint(userIDParam)
	if err != nil {
		logger.Error("Invalid userID format", zap.String("userID", userIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid userID format", "", "userID must be a number")
	}

	userConnectionsFollowers, err := h.socialService.GetUserConnectionsFollowers(ctx, userID, mainutils.StrPtrOrNil(cursorID))
	if err != nil {
		logger.Error("Failed to get user connections", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve user connection data", "", err.Error())
	}

	var nextCursorID *string
	if len(userConnectionsFollowers.Followers) > 0 {
		lastFollower := userConnectionsFollowers.Followers[len(userConnectionsFollowers.Followers)-1]
		nextCursorID = mainutils.UintToStrPtr(lastFollower.FollowID)
	}

	mappedData := fiber.Map{
		"followers":  userConnectionsFollowers.Followers,
		"nextCursor": nextCursorID,
	}

	return response.ResponseSuccess(c, 200, "User connection data retrieved successfully", "data", mappedData)
}

func (h *SocialHandler) GetUserConnectionsFollowingHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()

	userIDParam := c.Params("userID")
	cursorID := c.Query("cursorID")

	userID, err := mainutils.StringToUint(userIDParam)
	if err != nil {
		logger.Error("Invalid userID format", zap.String("userID", userIDParam), zap.Error(err))
		return response.ResponseError(c, 400, "Invalid userID format", "", "userID must be a number")
	}

	userConnections, err := h.socialService.GetUserConnectionsFollowing(ctx, userID, mainutils.StrPtrOrNil(cursorID))
	if err != nil {
		logger.Error("Failed to get user connections", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Failed to retrieve user connection data", "", err.Error())
	}

	var nextCursorID *string
	if len(userConnections.Following) > 0 {
		lastFollowing := userConnections.Following[len(userConnections.Following)-1]
		nextCursorID = mainutils.UintToStrPtr(lastFollowing.FollowID)
	}

	mappedData := fiber.Map{
		"following":  userConnections.Following,
		"nextCursor": nextCursorID,
	}

	return response.ResponseSuccess(c, 200, "User connection data retrieved successfully", "data", mappedData)
}
