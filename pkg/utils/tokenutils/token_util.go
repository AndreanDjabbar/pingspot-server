package tokenutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	mainutils "pingspot/pkg/utils/mainUtils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashString(txt string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(txt), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckHashString(txt, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(txt))
	return err == nil
}

func HashSHA256String(txt string) string {
	hash := sha256.Sum256(([]byte(txt)))
	return fmt.Sprintf("%x", hash)
}

func CheckHashSHA256String(txt, hash string) bool {
	hashedTxt := HashSHA256String(txt)
	return hashedTxt == hash
}

func GenerateRandomCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be greater than 0")
	}

	characters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	code := make([]byte, length)
	for i := range code {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(characters))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		code[i] = characters[index.Int64()]
	}
	return string(code), nil
}

func GenerateJWT(userID uint, JWTSecret []byte, email, username, fullName string) (string, error) {
	if len(JWTSecret) == 0 {
		return "", fmt.Errorf("JWT secret cannot be empty")
	}

	claims := jwt.MapClaims{
		"user_id":   userID,
		"email":     email,
		"username":  username,
		"full_name": fullName,
		"exp":       time.Now().Add(time.Hour * 10).Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

func ValidateRefreshToken(refreshToken string) (jwt.MapClaims, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is empty")
	}

	publicKeyPath := mainutils.GetKeyPath("public.pem")
	publicKey, err := LoadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	parsedToken, err := jwt.Parse(refreshToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	tokenType, ok := claims["token_type"].(string)
	if !ok || tokenType != "refresh" {
		return nil, fmt.Errorf("invalid refresh token type")
	}

	return claims, nil
}

func GetJWTClaims(c *fiber.Ctx) (jwt.MapClaims, error) {
	claims := c.Locals("claims")
	if claims != nil {
		if jwtClaims, ok := claims.(jwt.MapClaims); ok {
			return jwtClaims, nil
		}
	}

	token := c.Locals("token")
	if token == nil {
		return nil, fmt.Errorf("no JWT token found in context")
	}

	jwtToken, ok := token.(*jwt.Token)
	if !ok {
		return nil, fmt.Errorf("invalid JWT token type")
	}

	if jwtClaims, ok := jwtToken.Claims.(jwt.MapClaims); ok && jwtToken.Valid {
		return jwtClaims, nil
	}
	return nil, fmt.Errorf("invalid JWT token")
}

func ParseJWT(tokenString string, JWTSecret []byte) (jwt.MapClaims, error) {
	if len(JWTSecret) == 0 {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	publicKey, ok := pubInterface.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not RSA public key")
	}

	return publicKey, nil
}

func GenerateAccessToken(userID, sessionID uint, email, username, fullName string) string {
	privateKeyPath := mainutils.GetKeyPath("private.pem")
	privateKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		panic("Failed to load private key: " + err.Error())
	}

	claims := jwt.MapClaims{
		"user_id":    userID,
		"session_id": sessionID,
		"email":      email,
		"username":   username,
		"full_name":  fullName,
		"exp":        time.Now().Add(20 * time.Minute).Unix(),
		"iat":        time.Now().Unix(),
		"token_type": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		panic("Failed to sign token: " + err.Error())
	}
	return signedToken
}

func GenerateRefreshToken(userID uint, refreshTokenID string) string {
	privateKeyPath := mainutils.GetKeyPath("private.pem")
	privateKey, err := LoadPrivateKey(privateKeyPath)
	if err != nil {
		panic("Failed to load private key: " + err.Error())
	}

	claims := jwt.MapClaims{
		"user_id":          userID,
		"refresh_token_id": refreshTokenID,
		"exp":              time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat":              time.Now().Unix(),
		"token_type":       "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		panic("Failed to sign token: " + err.Error())
	}
	return signedToken
}
