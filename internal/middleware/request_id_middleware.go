package middleware

import (
	contextutils "pingspot/pkg/utils/contextUtils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := uuid.New().String()
		c.Locals("RequestID", requestID)

		ctx := contextutils.SetRequestIDInContext(c.Context(), requestID)
		c.SetUserContext(ctx)

		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}
