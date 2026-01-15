package middleware

import (
	"errors"
	"pingspot/pkg/utils/response"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/context"
)

func TimeoutMiddleware(d time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), d)
		defer cancel()

		c.SetUserContext(ctx)
		done := make(chan error, 1)

		go func() {
			done <- c.Next()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return response.ResponseError(c, 408, "Request timeout exceeded", "error", "Permintaan memakan waktu terlalu lama untuk diproses")
			}
			return ctx.Err()
		}
	}
}
