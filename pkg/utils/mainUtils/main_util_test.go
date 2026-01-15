package mainutils

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetClientIP(t *testing.T) {
	app := fiber.New()

	t.Run("should get IP from X-Forwarded-For header", func(t *testing.T) {
		var capturedIP string
		app.Get("/test", func(c *fiber.Ctx) error {
			capturedIP = GetClientIP(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", capturedIP)
	})

	t.Run("should get IP from X-Real-IP header", func(t *testing.T) {
		var capturedIP string
		app.Get("/test2", func(c *fiber.Ctx) error {
			capturedIP = GetClientIP(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test2", nil)
		req.Header.Set("X-Real-IP", "192.168.1.2")

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.2", capturedIP)
	})

	t.Run("should get IP from CF-Connecting-IP header", func(t *testing.T) {
		var capturedIP string
		app.Get("/test3", func(c *fiber.Ctx) error {
			capturedIP = GetClientIP(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test3", nil)
		req.Header.Set("CF-Connecting-IP", "192.168.1.3")

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.3", capturedIP)
	})

	t.Run("should prioritize X-Forwarded-For over other headers", func(t *testing.T) {
		var capturedIP string
		app.Get("/test4", func(c *fiber.Ctx) error {
			capturedIP = GetClientIP(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test4", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Set("X-Real-IP", "192.168.1.2")
		req.Header.Set("CF-Connecting-IP", "192.168.1.3")

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, "192.168.1.1", capturedIP)
	})
}

func TestGetHTTPClientIP(t *testing.T) {
	t.Run("should get IP from X-Forwarded-For header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")

		ip := GetHTTPClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("should get IP from X-Real-IP header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "192.168.1.2")

		ip := GetHTTPClientIP(req)
		assert.Equal(t, "192.168.1.2", ip)
	})

	t.Run("should get IP from CF-Connecting-IP header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("CF-Connecting-IP", "192.168.1.3")

		ip := GetHTTPClientIP(req)
		assert.Equal(t, "192.168.1.3", ip)
	})

	t.Run("should return RemoteAddr when no headers present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		ip := GetHTTPClientIP(req)
		assert.NotEmpty(t, ip)
	})
}

func TestGetUserAgent(t *testing.T) {
	app := fiber.New()

	t.Run("should get User-Agent from request", func(t *testing.T) {
		var capturedUA string
		app.Get("/test", func(c *fiber.Ctx) error {
			capturedUA = GetUserAgent(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Test Browser)")

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, "Mozilla/5.0 (Test Browser)", capturedUA)
	})

	t.Run("should return empty string when User-Agent not present", func(t *testing.T) {
		var capturedUA string
		app.Get("/test2", func(c *fiber.Ctx) error {
			capturedUA = GetUserAgent(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test2", nil)

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, "", capturedUA)
	})
}

func TestGetHTTPUserAgent(t *testing.T) {
	t.Run("should get User-Agent from request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Test Browser)")

		ua := GetHTTPUserAgent(req)
		assert.Equal(t, "Mozilla/5.0 (Test Browser)", ua)
	})

	t.Run("should return empty string when User-Agent not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		ua := GetHTTPUserAgent(req)
		assert.Equal(t, "", ua)
	})
}

func TestGetDeviceInfo(t *testing.T) {
	app := fiber.New()

	t.Run("should return formatted device info", func(t *testing.T) {
		var capturedInfo string
		app.Get("/test", func(c *fiber.Ctx) error {
			capturedInfo = GetDeviceInfo(c)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1")
		req.Header.Set("User-Agent", "Mozilla/5.0")

		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Contains(t, capturedInfo, "IP: 192.168.1.1")
		assert.Contains(t, capturedInfo, "UA: Mozilla/5.0")
	})
}

func TestGetKeyPath(t *testing.T) {
	t.Run("should return correct key path", func(t *testing.T) {
		filename := "test.pem"
		path := GetKeyPath(filename)

		assert.Contains(t, path, "keys")
		assert.Contains(t, path, filename)
	})

	t.Run("should handle different filenames", func(t *testing.T) {
		filenames := []string{"public.pem", "private.pem", "certificate.crt"}

		for _, filename := range filenames {
			path := GetKeyPath(filename)
			assert.Contains(t, path, filename)
			assert.Contains(t, path, "keys")
		}
	})
}

func TestRenderEmailTemplate(t *testing.T) {
	t.Run("should return error for verification email without link", func(t *testing.T) {
		data := EmailData{
			To:            "test@example.com",
			RecipientName: "Test User",
			EmailType:     EmailTypeVerification,
			TemplateData:  map[string]any{},
		}

		_, err := RenderEmailTemplate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "verification link is required")
	})

	t.Run("should return error for password reset email without link", func(t *testing.T) {
		data := EmailData{
			To:            "test@example.com",
			RecipientName: "Test User",
			EmailType:     EmailTypePasswordReset,
			TemplateData:  map[string]any{},
		}

		_, err := RenderEmailTemplate(data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "reset link is required")
	})

	t.Run("should render verification email template", func(t *testing.T) {
		templateHTML := "<html><body>Hello {{.UserName}}, verify: {{.VerificationLink}}</body></html>"
		data := EmailData{
			To:            "test@example.com",
			RecipientName: "Test User",
			EmailType:     EmailTypeVerification,
			BodyTempate:   templateHTML,
			TemplateData: map[string]any{
				"VerificationLink": "https://example.com/verify?token=123",
			},
		}

		html, err := RenderEmailTemplate(data)
		require.NoError(t, err)
		assert.Contains(t, html, "Test User")
		assert.Contains(t, html, "https://example.com/verify?token=123")
	})

	t.Run("should render password reset email template", func(t *testing.T) {
		templateHTML := "<html><body>Hello {{.UserName}}, reset: {{.ResetLink}}</body></html>"
		data := EmailData{
			To:            "test@example.com",
			RecipientName: "Test User",
			EmailType:     EmailTypePasswordReset,
			BodyTempate:   templateHTML,
			TemplateData: map[string]any{
				"ResetLink": "https://example.com/reset?token=456",
			},
		}

		html, err := RenderEmailTemplate(data)
		require.NoError(t, err)
		assert.Contains(t, html, "Test User")
		assert.Contains(t, html, "https://example.com/reset?token=456")
	})
}

// Benchmark tests
func BenchmarkGetClientIP(b *testing.B) {
	app := fiber.New()
	app.Get("/bench", func(c *fiber.Ctx) error {
		_ = GetClientIP(c)
		return c.SendStatus(200)
	})

	req := httptest.NewRequest("GET", "/bench", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = app.Test(req)
	}
}

func BenchmarkGetHTTPClientIP(b *testing.B) {
	req := httptest.NewRequest("GET", "/bench", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetHTTPClientIP(req)
	}
}
