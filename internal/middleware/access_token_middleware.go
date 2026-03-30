package middleware

import (
	"context"
	"fmt"
	"pingspot/internal/infrastructure/cache"
	mainutils "pingspot/pkg/utils/mainUtils"
	"pingspot/pkg/utils/response"
	"pingspot/pkg/utils/tokenutils"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

        if tokenType, _ := claims["token_type"].(string); tokenType != "access" {
            return response.ResponseError(c, 401, "Token tidak valid", "", "Tipe token tidak sesuai")
        }

        sessionIDFloat, ok := claims["session_id"].(float64)
        if !ok {
            return response.ResponseError(c, 401, "Token tidak valid", "", "Session ID tidak ditemukan pada token")
        }

        sessionID := uint(sessionIDFloat)

        redisClient := cache.GetRedis()
        userSessionKey := fmt.Sprintf("session:%v", sessionID)
        storedUserID, err := redisClient.Get(context.Background(), userSessionKey).Result()
        if err != nil || storedUserID == "" || storedUserID != fmt.Sprintf("%v", claims["user_id"]) {
            return response.ResponseError(c, 401, "Sesi tidak valid", "", "Sesi pengguna tidak ditemukan atau sudah tidak berlaku")
        }
        
        c.Locals("token", parsedToken)
        c.Locals("claims", claims)

        return c.Next()
    }
}
