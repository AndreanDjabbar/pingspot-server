package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"pingspot/internal/domain/authService/dto"
	"pingspot/internal/domain/authService/service"
	"pingspot/internal/domain/authService/util"
	"pingspot/internal/domain/authService/validation"
	"pingspot/internal/infrastructure/cache"
	"pingspot/internal/infrastructure/database"
	apperror "pingspot/pkg/apperror"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"
	mainutils "pingspot/pkg/utils/mainUtils"
	"pingspot/pkg/utils/response"
	"pingspot/pkg/utils/tokenutils"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth/gothic"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func getAccessTokenAge() int {
	ageStr := env.AccessTokenAge()
	age, err := strconv.Atoi(ageStr)
	if err != nil || age <= 0 {
		return 1200
	}
	return age
}

func getRefreshTokenAge() int {
	ageStr := env.RefreshTokenAge()
	age, err := strconv.Atoi(ageStr)
	if err != nil || age <= 0 {
		return 604800
	}
	return age
}

func (h *AuthHandler) RegisterHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()

	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatRegisterValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}
	db := database.GetPostgresDB()
	user, err := h.authService.Register(ctx, db, req, false)
	if err != nil {
		logger.Error("Registration failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Registrasi gagal", "", err.Error())
	}

	newUser := map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}

	randomCode1, err := tokenutils.GenerateRandomCode(150)
	if err != nil {
		logger.Error("Failed to generate random code", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal membuat kode acak", "", err.Error())
	}
	randomCode2, err := tokenutils.GenerateRandomCode(150)
	if err != nil {
		logger.Error("Failed to generate random code", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal membuat kode acak", "", err.Error())
	}
	verificationLink := fmt.Sprintf("%s/auth/verification?code1=%s&userId=%d&code2=%s", env.ClientURL(), randomCode1, user.ID, randomCode2)

	redisClient := cache.GetRedis()

	linkData := map[string]string{
		"link1": randomCode1,
		"link2": randomCode2,
	}
	linkJSON, err := json.Marshal(linkData)
	if err != nil {
		logger.Error("Failed to marshal verification link data", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal menyimpan kode verifikasi", "", err.Error())
	}

	redisKey := fmt.Sprintf("link:%d", newUser["id"])
	err = redisClient.Set(c.Context(), redisKey, linkJSON, 300*time.Second).Err()
	if err != nil {
		logger.Error("Failed to save verification link to Redis", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal menyimpan kode verifikasi ke Redis", "", err.Error())
	}

	go util.SendVerificationEmail(newUser["email"].(string), newUser["username"].(string), verificationLink)

	logger.Info("User registered successfully", zap.String("user_id", fmt.Sprintf("%d", user.ID)))

	return response.ResponseSuccess(c, 200, "Registrasi berhasil. Silahkan cek email anda untuk verifikasi akun", "data", nil)
}

func (h *AuthHandler) VerificationHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	code1 := c.Query("code1")
	userId := c.Query("userId")
	code2 := c.Query("code2")

	if code1 == "" || userId == "" || code2 == "" {
		return response.ResponseError(c, 400, "Parameter tidak lengkap", "", nil)
	}

	redisClient := cache.GetRedis()
	redisKey := fmt.Sprintf("link:%s", userId)
	linkData, err := redisClient.Get(c.Context(), redisKey).Result()
	if err != nil {
		var errorMsg string
		if err == redis.Nil {
			errorMsg = "Link verifikasi tidak ditemukan atau sudah kadaluarsa"
		} else {
			errorMsg = "Gagal mendapatkan kode verifikasi"
		}
		logger.Error("Failed to get verification link from Redis", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal mendapatkan kode verifikasi", "", errorMsg)
	}

	var link map[string]string
	if err := json.Unmarshal([]byte(linkData), &link); err != nil {
		logger.Error("Failed to unmarshal verification link data", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal memproses link verifikasi", "", err.Error())
	}

	if link["link1"] != code1 || link["link2"] != code2 {
		return response.ResponseError(c, 400, "link verifikasi tidak valid", "", "Link verifikasi tidak valid")
	}

	userIdUint, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		logger.Error("Invalid user ID format", zap.Error(err))
		return response.ResponseError(c, 400, "ID pengguna tidak valid", "", err.Error())
	}

	user, err := h.authService.VerifyUser(ctx, uint(userIdUint))
	if err != nil {
		logger.Error("Verification failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Verifikasi gagal", "", err.Error())
	}

	if err := redisClient.Del(c.Context(), redisKey).Err(); err != nil {
		logger.Error("Failed to delete verification link from Redis", zap.Error(err))
	}

	return response.ResponseSuccess(c, 200, "Akun berhasil diverifikasi", "data", dto.VerificationResponse{
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
	})
}

func (h *AuthHandler) LoginHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	db := database.GetPostgresDB()
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}

	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatLoginValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	userIP := mainutils.GetClientIP(c)
	userAgent := mainutils.GetUserAgent(c)

	req.IPAddress = userIP
	req.UserAgent = userAgent

	_, accessToken, refreshToken, err := h.authService.Login(ctx, db, req)
	if err != nil {
		logger.Error("Login failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 401, "Login gagal", "", err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Domain:   "",
		Path:     "/",
		MaxAge:   getAccessTokenAge(),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Domain:   "",
		Path:     "/",
		MaxAge:   getRefreshTokenAge(),
	})

	return response.ResponseSuccess(c, 200, "Login berhasil", "data", dto.LoginResponse{
		AccessToken: accessToken,
		ExpiresIn:   int64(getAccessTokenAge()),
	})
}

func (h *AuthHandler) OAuthLoginHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("OAUTH LOGIN HANDLER", zap.String("provider", provider))
		r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
		gothic.BeginAuthHandler(w, r)
	}
}

func (h *AuthHandler) GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
	h.OAuthLoginHandler("google")(w, r)
}

func (h *AuthHandler) OAuthCallbackHandler(provider string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("OAUTH CALLBACK HANDLER", zap.String("provider", provider))
		r = r.WithContext(context.WithValue(context.Background(), "provider", provider))
		user, err := gothic.CompleteUserAuth(w, r)
		if err != nil {
			logger.Error("OAuth authentication failed", zap.String("provider", provider), zap.Error(err))
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
			return
		}
		email := user.Email
		fullName := user.Name
		nickName := user.RawData["given_name"].(string)
		if nickName == "" {
			nickName = user.RawData["name"].(string)
		}
		providerId := user.RawData["id"].(string)
		logger.Info("OAuth user authenticated", zap.String("provider", provider), zap.String("email", email), zap.String("name", fullName))

		ctx := r.Context()
		existingUser, err := h.authService.GetUserByEmail(ctx, email)
		if err != nil {
			logger.Error("Error retrieving user by email", zap.Error(err))
			http.Error(w, "Terdapat masalah", http.StatusNotFound)
			return
		}

		if existingUser == nil {
			newUser := dto.RegisterRequest{
				Username:   nickName,
				Email:      email,
				FullName:   fullName,
				Provider:   provider,
				ProviderID: &providerId,
			}
			db := database.GetPostgresDB()
			createdUser, err := h.authService.Register(ctx, db, newUser, true)
			if err != nil {
				logger.Error("Error registering new user", zap.String("provider", provider), zap.Error(err))
				http.Error(w, "Terdapat masalah saat registrasi", http.StatusInternalServerError)
				return
			}
			logger.Info("New user registered", zap.String("provider", provider), zap.String("user_id", fmt.Sprintf("%d", createdUser.ID)))
			existingUser = createdUser
		}

		db := database.GetPostgresDB()

		var loginReq dto.LoginRequest
		loginReq.Email = existingUser.Email
		loginReq.Password = ""

		userIP := mainutils.GetHTTPClientIP(r)
		userAgent := mainutils.GetHTTPUserAgent(r)

		loginReq.IPAddress = userIP
		loginReq.UserAgent = userAgent

		_, accessToken, refreshToken, err := h.authService.Login(ctx, db, loginReq)
		if err != nil {
			logger.Error("Login failed for OAuth user", zap.String("provider", provider), zap.Error(err))
			http.Error(w, "Terdapat masalah saat login", http.StatusInternalServerError)
			return
		}

		cookieAccess := &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Domain:   "",
			Path:     "/",
			MaxAge:   getAccessTokenAge(),
		}
		http.SetCookie(w, cookieAccess)

		cookieRefresh := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
			Domain:   "",
			Path:     "/",
			MaxAge:   getRefreshTokenAge(),
		}
		http.SetCookie(w, cookieRefresh)

		http.Redirect(w, r, fmt.Sprintf("%s/auth/%s?status=%d", env.ClientURL(), provider, http.StatusAccepted), http.StatusFound)
	}
}

func (h *AuthHandler) GoogleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	h.OAuthCallbackHandler("google")(w, r)
}

func (h *AuthHandler) ForgotPasswordEmailVerificationHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dto.ForgotPasswordEmailVerificationRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatForgotPasswordEmailVerificationValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	user, err := h.authService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to get user by email", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan pengguna", "", err.Error())
	}
	if user != nil {
		redisClient := cache.GetRedis()
		verificationCode, err := tokenutils.GenerateRandomCode(200)
		if err != nil {
			logger.Error("Failed to generate verification code", zap.Error(err))
			return response.ResponseError(c, 500, "Gagal membuat kode verifikasi", "", err.Error())
		}
		verificationLink := fmt.Sprintf("%s/auth/forgot-password/verification?code=%s&email=%s", env.ClientURL(), verificationCode, req.Email)
		redisKey := fmt.Sprintf("forgot_password:%s", req.Email)
		err = redisClient.Set(c.Context(), redisKey, verificationCode, 300*time.Second).Err()
		if err != nil {
			logger.Error("Failed to save verification code to Redis", zap.Error(err))
			return response.ResponseError(c, 500, "Gagal menyimpan kode verifikasi ke Redis", "", err.Error())
		}

		go util.SendPasswordResetEmail(req.Email, req.Email, verificationLink)
	}
	return response.ResponseSuccess(c, 200, "Silahkan cek email anda untuk verifikasi pengaturan ulang kata sandi", "data", nil)
}

func (h *AuthHandler) ForgotPasswordLinkVerificationHandler(c *fiber.Ctx) error {
	code := c.Query("code")
	email := c.Query("email")

	if code == "" || email == "" {
		return response.ResponseError(c, 400, "Parameter tidak lengkap", "", nil)
	}

	redisClient := cache.GetRedis()
	redisKey := fmt.Sprintf("forgot_password:%s", email)
	storedCode, err := redisClient.Get(c.Context(), redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return response.ResponseError(c, 400, "Link verifikasi tidak ditemukan atau sudah kadaluarsa", "", nil)
		}
		logger.Error("Failed to get verification code from Redis", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal mendapatkan kode verifikasi", "", err.Error())
	}

	if storedCode != code {
		return response.ResponseError(c, 400, "Link verifikasi tidak valid", "", nil)
	}

	return response.ResponseSuccess(c, 200, "Link verifikasi berhasil", "data", dto.ForgotPasswordLinkVerificationResponse{
		Email: email,
	})
}

func (h *AuthHandler) RefreshTokenHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return response.ResponseError(c, 401, "Refresh token tidak ditemukan", "", "Anda harus login terlebih dahulu")
	}

	newAccessToken, newRefreshToken, err := h.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		logger.Error("Failed to refresh tokens", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memperbarui token", "", err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Domain:   "",
		Path:     "/",
		MaxAge:   getAccessTokenAge(),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Domain:   "",
		Path:     "/",
		MaxAge:   getRefreshTokenAge(),
	})

	return response.ResponseSuccess(c, 200, "Token berhasil diperbarui", "data", dto.RefreshTokenResponse{
		AccessToken: newAccessToken,
		ExpiresIn:   int64(getAccessTokenAge()),
	})
}

func (h *AuthHandler) ForgotPasswordResetPasswordHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	var req dto.ForgotPasswordResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		logger.Error("Failed to parse request body", zap.Error(err))
		return response.ResponseError(c, 400, "Format body request tidak valid", "", err.Error())
	}
	if err := validation.Validate.Struct(req); err != nil {
		errors := validation.FormatForgotPasswordResetPasswordValidationErrors(err)
		logger.Error("Validation failed", zap.Error(err))
		return response.ResponseError(c, 400, "Validasi gagal", "errors", errors)
	}

	user, err := h.authService.GetUserByEmail(ctx, req.Email)
	if err != nil {
		logger.Error("Failed to get user by email", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mendapatkan pengguna", "", err.Error())
	}
	if user == nil {
		return response.ResponseError(c, 404, "Pengguna tidak ditemukan", "", "Email tidak terdaftar")
	}

	hashNewPassword, err := tokenutils.HashString(req.Password)
	if err != nil {
		logger.Error("Failed to hash new password", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal mengenkripsi kata sandi baru", "", err.Error())
	}
	user.Password = &hashNewPassword

	updatedUser, err := h.authService.UpdateUserByEmail(ctx, req.Email, user)
	if err != nil {
		logger.Error("Failed to update user password", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memperbarui kata sandi", "", err.Error())
	}
	if updatedUser == nil {
		return response.ResponseError(c, 404, "Pengguna tidak ditemukan", "", "Email tidak terdaftar")
	}

	return response.ResponseSuccess(c, 200, "Kata sandi berhasil diperbarui. Silahkan masuk dengan identitas terbaru anda", "data", nil)
}

func (h *AuthHandler) LogoutHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return response.ResponseError(c, 401, "Refresh token tidak ditemukan", "", "Anda harus login terlebih dahulu")
	}

	refreshTokenClaims, err := tokenutils.ValidateRefreshToken(refreshToken)
	if err != nil {
		logger.Error("Failed to validate refresh token", zap.Error(err))
		return response.ResponseError(c, 401, "Refresh token tidak valid", "", "Token tidak dapat diverifikasi")
	}

	refreshTokenID := refreshTokenClaims["refresh_token_id"].(string)
	userID := uint(refreshTokenClaims["user_id"].(float64))

	if err := h.authService.Logout(ctx, userID, refreshTokenID); err != nil {
		logger.Error("Failed to logout user", zap.Error(err))
		return response.ResponseError(c, 500, "Gagal logout", "", err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Domain:   "",
		MaxAge:   -1,
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Domain:   "",
		MaxAge:   -1,
		Path:     "/",
	})

	return response.ResponseSuccess(c, 200, "Logout berhasil", "data", nil)
}
