package handler

import (
	"context"
	"fmt"
	"net/http"
	"pingspot/internal/domain/auth_service/dto"
	"pingspot/internal/domain/auth_service/service"
	"pingspot/internal/domain/auth_service/validation"
	apperror "pingspot/pkg/app_error"
	"pingspot/pkg/logger"
	env "pingspot/pkg/utils/env_util"
	mainutils "pingspot/pkg/utils/main_util"
	response "pingspot/pkg/utils/response_util"
	tokenutils "pingspot/pkg/utils/token_util"
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth/gothic"
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
	user, err := h.authService.Register(ctx, req, false)
	if err != nil {
		logger.Error("Registration failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Registrasi gagal", "", err.Error())
	}

	if err := h.authService.SendRegistrationVerificationEmail(ctx, user); err != nil {
		logger.Error("Failed to send verification email", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal mengirim email verifikasi", "", err.Error())
	}

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

	userIdUint, err := strconv.ParseUint(userId, 10, 32)
	if err != nil {
		logger.Error("Invalid user ID format", zap.Error(err))
		return response.ResponseError(c, 400, "ID pengguna tidak valid", "", err.Error())
	}

	user, err := h.authService.VerifyRegistrationCode(ctx, code1, code2, uint(userIdUint))
	if err != nil {
		logger.Error("Verification failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Verifikasi gagal", "", err.Error())
	}

	return response.ResponseSuccess(c, 200, "Akun berhasil diverifikasi", "data", dto.VerificationResponse{
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
	})
}

func (h *AuthHandler) LoginHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
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

	_, accessToken, refreshToken, err := h.authService.Login(ctx, req)
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
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   "",
		Path:     "/",
		MaxAge:   getAccessTokenAge(),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
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

		userIP := mainutils.GetHTTPClientIP(r)
		userAgent := mainutils.GetHTTPUserAgent(r)

		ctx := r.Context()
		accessToken, refreshToken, err := h.authService.HandleOAuthCallback(ctx, email, fullName, nickName, user.Name, providerId, provider, userIP, userAgent)
		if err != nil {
			logger.Error("OAuth callback handler failed", zap.String("provider", provider), zap.Error(err))
			if appErr, ok := err.(*apperror.AppError); ok {
				http.Error(w, appErr.Message, appErr.StatusCode)
			} else {
				http.Error(w, "Terdapat masalah", http.StatusInternalServerError)
			}
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "access_token",
			Value:    accessToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   getAccessTokenAge(),
		})

		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			Path:     "/",
			MaxAge:   getRefreshTokenAge(),
		})

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

	err := h.authService.ForgotPasswordEmailVerification(ctx, req)
	if err != nil {
		logger.Error("Forgot password email verification failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memproses permintaan", "", err.Error())
	}

	return response.ResponseSuccess(c, 200, "Silahkan cek email anda untuk verifikasi pengaturan ulang kata sandi", "data", nil)
}

func (h *AuthHandler) ForgotPasswordLinkVerificationHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	code := c.Query("code")
	email := c.Query("email")

	if code == "" || email == "" {
		return response.ResponseError(c, 400, "Parameter tidak lengkap", "", nil)
	}

	if err := h.authService.VerifyForgotPasswordCode(ctx, code, email); err != nil {
		logger.Error("Forgot password code verification failed", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memverifikasi kode", "", err.Error())
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
			if appErr.StatusCode == 401 {
				tokenutils.ClearAuthCookies(c)
			}
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memperbarui token", "", err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   "",
		Path:     "/",
		MaxAge:   getAccessTokenAge(),
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
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

	if err := h.authService.ResetPassword(ctx, req.Email, req.Password); err != nil {
		logger.Error("Failed to reset password", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal memperbarui kata sandi", "", err.Error())
	}

	return response.ResponseSuccess(c, 200, "Kata sandi berhasil diperbarui. Silahkan masuk dengan identitas terbaru anda", "data", nil)
}

func (h *AuthHandler) LogoutHandler(c *fiber.Ctx) error {
	ctx := c.UserContext()
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return response.ResponseError(c, 401, "Refresh token tidak ditemukan", "", "Anda harus login terlebih dahulu")
	}

	if err := h.authService.Logout(ctx, refreshToken); err != nil {
		logger.Error("Failed to logout user", zap.Error(err))
		if appErr, ok := err.(*apperror.AppError); ok {
			return response.ResponseError(c, appErr.StatusCode, appErr.Message, "error_code", appErr.Code)
		}
		return response.ResponseError(c, 500, "Gagal logout", "", err.Error())
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   "",
		MaxAge:   -1,
		Path:     "/",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteLaxMode,
		Domain:   "",
		MaxAge:   -1,
		Path:     "/",
	})

	return response.ResponseSuccess(c, 200, "Logout berhasil", "data", nil)
}
