package middleware

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http/httptest"
	"os"
	"path/filepath"
	contextutils "pingspot/pkg/utils/context_util"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestKeys(t *testing.T) *rsa.PrivateKey {
	t.Helper()

	keysDir := filepath.Join("keys")
	err := os.MkdirAll(keysDir, 0755)
	require.NoError(t, err)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	privateKeyFile := filepath.Join(keysDir, "private.pem")
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}

	privateFile, err := os.Create(privateKeyFile)
	require.NoError(t, err)
	defer privateFile.Close()
	err = pem.Encode(privateFile, privateKeyPEM)
	require.NoError(t, err)

	publicKeyFile := filepath.Join(keysDir, "public.pem")
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	require.NoError(t, err)

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	publicFile, err := os.Create(publicKeyFile)
	require.NoError(t, err)
	defer publicFile.Close()
	err = pem.Encode(publicFile, publicKeyPEM)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = os.RemoveAll(keysDir)
	})

	return privateKey
}

func signToken(t *testing.T, key *rsa.PrivateKey, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signed, err := token.SignedString(key)
	require.NoError(t, err)
	return signed
}

func TestRequestIDMiddleware_SetsHeaderLocalsAndContext(t *testing.T) {
	app := fiber.New()
	app.Use(RequestIDMiddleware())
	app.Get("/", func(c *fiber.Ctx) error {
		requestID, ok := c.Locals("RequestID").(string)
		require.True(t, ok)
		_, err := uuid.Parse(requestID)
		require.NoError(t, err)
		assert.Equal(t, requestID, contextutils.GetRequestID(c.UserContext()))
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get("X-Request-ID"))
}

func TestTimeoutMiddleware_SuccessPath(t *testing.T) {
	app := fiber.New()
	app.Use(TimeoutMiddleware(100 * time.Millisecond))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestTimeoutMiddleware_TimeoutPath(t *testing.T) {
	app := fiber.New()
	app.Use(TimeoutMiddleware(5 * time.Millisecond))
	app.Get("/", func(c *fiber.Ctx) error {
		time.Sleep(30 * time.Millisecond)
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusRequestTimeout, resp.StatusCode)
}

func TestLoggingMiddleware_PropagatesHandlerError(t *testing.T) {
	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("RequestID", "test-request-id")
		return c.Next()
	})
	app.Use(LoggingMiddleware())

	testErr := errors.New("handler failed")
	app.Get("/", func(c *fiber.Ctx) error {
		return testErr
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestValidateAccessToken_NoToken(t *testing.T) {
	app := fiber.New()
	app.Use(ValidateAccessToken())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestValidateAccessToken_PublicKeyLoadFails(t *testing.T) {
	app := fiber.New()
	app.Use(ValidateAccessToken())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestValidateAccessToken_RejectsWrongTokenType(t *testing.T) {
	privateKey := writeTestKeys(t)
	claims := jwt.MapClaims{
		"token_type": "refresh",
		"user_id":    1,
		"session_id": 1,
		"exp":        time.Now().Add(5 * time.Minute).Unix(),
	}
	token := signToken(t, privateKey, claims)

	app := fiber.New()
	app.Use(ValidateAccessToken())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestValidateAccessToken_MissingSessionIDClaim(t *testing.T) {
	privateKey := writeTestKeys(t)
	claims := jwt.MapClaims{
		"token_type": "access",
		"user_id":    1,
		"exp":        time.Now().Add(5 * time.Minute).Unix(),
	}
	token := signToken(t, privateKey, claims)

	app := fiber.New()
	app.Use(ValidateAccessToken())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestValidateAccessToken_MissingUserIDClaim(t *testing.T) {
	privateKey := writeTestKeys(t)
	claims := jwt.MapClaims{
		"token_type": "access",
		"session_id": 1,
		"exp":        time.Now().Add(5 * time.Minute).Unix(),
	}
	token := signToken(t, privateKey, claims)

	app := fiber.New()
	app.Use(ValidateAccessToken())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	privateKey := writeTestKeys(t)
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := jwt.MapClaims{
		"token_type": "access",
		"session_id": 1,
		"user_id":    1,
		"exp":        time.Now().Add(5 * time.Minute).Unix(),
	}
	_ = privateKey
	token := signToken(t, otherKey, claims)

	app := fiber.New()
	app.Use(ValidateAccessToken())
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}

func TestTimeoutMiddleware_TimeoutBodyShape(t *testing.T) {
	app := fiber.New()
	app.Use(TimeoutMiddleware(5 * time.Millisecond))
	app.Get("/", func(c *fiber.Ctx) error {
		time.Sleep(30 * time.Millisecond)
		return nil
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var body map[string]any
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, false, body["success"])
	assert.Equal(t, "Request timeout exceeded", body["message"])
}
