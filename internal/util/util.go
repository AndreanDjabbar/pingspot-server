package util

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

type contextKey string

const requestIDKey contextKey = "request_id"

func SetRequestIDInContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

func GetClientIP(c *fiber.Ctx) string {
	if ip := c.Get("X-Forwarded-For"); ip != "" {
		for i := 0; i < len(ip); i++ {
			if ip[i] == ',' {
				return ip[:i]
			}
		}
		return ip
	}

	if ip := c.Get("X-Real-IP"); ip != "" {
		return ip
	}

	if ip := c.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	ip := c.IP()

	if ip == "::1" {
		ip = "127.0.0.1"
	}

	return ip
}

func GetHTTPClientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if idx := len(ip); idx > 0 {
			for i := 0; i < idx; i++ {
				if ip[i] == ',' {
					return ip[:i]
				}
			}
			return ip
		}
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}

func GetHTTPUserAgent(r *http.Request) string {
	return r.Header.Get("User-Agent")
}

func GetDeviceInfo(c *fiber.Ctx) string {
	return fmt.Sprintf("IP: %s, UA: %s", GetClientIP(c), GetUserAgent(c))
}

func GetKeyPath(filename string) string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "certs", filename)
}




