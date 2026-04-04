package middleware

import (
	"fmt"
	"pingspot/internal/domain/user_service/repository"
	"pingspot/internal/infrastructure/cache"
	"pingspot/internal/infrastructure/database"
	"pingspot/pkg/logger"
	mainutils "pingspot/pkg/utils/main_util"
	response "pingspot/pkg/utils/response_util"
	tokenutils "pingspot/pkg/utils/token_util"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func ValidateAccessToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token = authHeader[7:]
		} else {
			token = c.Cookies("access_token")
		}

		if token == "" {
			return response.ResponseError(c, 401, "Token tidak ditemukan", "", "Anda harus login terlebih dahulu")
		}

		publicKeyPath := mainutils.GetKeyPath("public.pem")
		publicKey, err := tokenutils.LoadPublicKey(publicKeyPath)
		if err != nil {
			return response.ResponseError(c, 500, "Gagal memuat kunci publik", "", err.Error())
		}

		parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return publicKey, nil
		})
		if err != nil || !parsedToken.Valid {
			return response.ResponseError(c, 401, "Token tidak valid", "", "Token tidak dapat diverifikasi")
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return response.ResponseError(c, 401, "Token tidak valid", "", "Claims tidak dapat dibaca")
		}

		tokenType, _ := claims["token_type"].(string)
		if tokenType != "access" {
			return response.ResponseError(c, 401, "Token tidak valid", "", "Tipe token tidak sesuai")
		}

		sessionIDFloat, ok := claims["session_id"].(float64)
		if !ok {
			return response.ResponseError(c, 401, "Token tidak valid", "", "Session ID tidak ditemukan pada token")
		}
		sessionID := uint(sessionIDFloat)

		userIDFloat, ok := claims["user_id"].(float64)
		if !ok {
			return response.ResponseError(c, 401, "Token tidak valid", "", "User ID tidak ditemukan pada token")
		}
		userID := uint(userIDFloat)

		ctx := c.UserContext()
		redisClient := cache.GetRedis()
		sessionKey := fmt.Sprintf("session:%d", sessionID)

		storedUserID, err := redisClient.Get(ctx, sessionKey).Result()
		if err == nil {
			if storedUserID != fmt.Sprintf("%d", userID) {
				tokenutils.ClearAuthCookies(c)
				return response.ResponseError(c, 401, "Sesi tidak valid", "", "Sesi tidak sesuai dengan user")
			}

			c.Locals("token", parsedToken)
			c.Locals("claims", claims)
			return c.Next()
		}

		db := database.GetPostgresDB()
		userSessionRepo := repository.NewUserSessionRepository(db)

		userSession, err := userSessionRepo.GetByID(ctx, sessionID)
		if err != nil {
			tokenutils.ClearAuthCookies(c)
			return response.ResponseError(c, 401, "Sesi tidak valid", "", "Sesi pengguna tidak ditemukan atau sudah tidak berlaku")
		}

		if !userSession.IsActive {
			tokenutils.ClearAuthCookies(c)
			return response.ResponseError(c, 401, "Sesi tidak valid", "", "Sesi pengguna sudah tidak aktif")
		}

		if userSession.ExpiresAt < time.Now().Unix() {
			tokenutils.ClearAuthCookies(c)
			return response.ResponseError(c, 401, "Sesi kedaluwarsa", "", "Sesi pengguna sudah kedaluwarsa")
		}

		if userSession.UserID != userID {
			tokenutils.ClearAuthCookies(c)
			return response.ResponseError(c, 401, "Sesi tidak valid", "", "Sesi tidak sesuai dengan user")
		}

		ttl := time.Until(time.Unix(userSession.ExpiresAt, 0))
		if ttl > 0 {
			if err := redisClient.Set(ctx, sessionKey, fmt.Sprintf("%d", userID), ttl).Err(); err != nil {
				logger.Warn("Failed to rebuild session cache in Redis", zap.Error(err))
			}
		}

		c.Locals("token", parsedToken)
		c.Locals("claims", claims)

		return c.Next()
	}
}