package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"pingspot/internal/model"
	"pingspot/internal/repository"
	"pingspot/internal/util"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const idempotencyKeyHeader = "Idempotency-Key"

func IdempotencyMiddleware(repo repository.IdempotencyRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := strings.TrimSpace(c.Get(idempotencyKeyHeader))
		fmt.Println("Idempotency Key:", key) // Debugging line to print the Idempotency-Key
		fmt.Println("Request Method:", c.Method()) // Debugging line to print the request method
		if key == "" {
			return util.ResponseError(c, fiber.StatusBadRequest, "Missing Idempotency-Key", "", "Idempotency-Key header is required")
		}
		if len(key) > 255 {
			return util.ResponseError(c, fiber.StatusBadRequest, "Invalid Idempotency-Key", "", "Idempotency-Key must be 255 characters or fewer")
		}

		userID, ok := c.Locals("userID").(uint)
		if !ok || userID == 0 {
			return util.ResponseError(c, fiber.StatusUnauthorized, "User not authenticated", "", "Unable to identify the authenticated user")
		}
		fmt.Println("User ID:", userID) // Debugging line to print the user ID

		method := c.Method()
		route := c.Route().Path
		hash := sha256.Sum256(c.Body())
		requestHash := hex.EncodeToString(hash[:])

		record, err := repo.Get(c.UserContext(), userID, key, method, route)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return util.ResponseError(c, fiber.StatusInternalServerError, "Idempotency lookup failed", "", err.Error())
		}

		if record != nil {
			if record.RequestHash != requestHash {
				return util.ResponseError(c, fiber.StatusConflict, "Idempotency-Key already used", "", "Use a new key for a different request")
			}
			if record.Status == "PROCESSING" {
				c.Set("Retry-After", "2")
				return util.ResponseError(c, fiber.StatusConflict, "Request is still processing", "", "Retry with the same Idempotency-Key")
			}
			if record.ResponseStatus == nil {
				return util.ResponseError(c, fiber.StatusInternalServerError, "Invalid idempotency record", "", "The stored response status is missing")
			}
			c.Status(*record.ResponseStatus)
			c.Response().Header.SetContentType(fiber.MIMEApplicationJSON)
			return c.Send(*record.ResponseBody)
		}

		record = &model.IdempotencyRecord{
			UserID:         userID,
			IdempotencyKey: key,
			Method:         method,
			Route:          route,
			RequestHash:    requestHash,
			Status:         "PROCESSING",
		}
		if err := repo.Create(c.UserContext(), record); err != nil {
			if !errors.Is(err, repository.ErrIdempotencyRecordConflict) {
				return util.ResponseError(c, fiber.StatusInternalServerError, "Idempotency reservation failed", "", err.Error())
			}

			existing, getErr := repo.Get(c.UserContext(), userID, key, method, route)
			if getErr != nil {
				return util.ResponseError(c, fiber.StatusInternalServerError, "Idempotency lookup failed", "", getErr.Error())
			}
			if existing.RequestHash != requestHash {
				return util.ResponseError(c, fiber.StatusConflict, "Idempotency-Key already used", "", "Use a new key for a different request")
			}
			c.Set("Retry-After", "2")
			return util.ResponseError(c, fiber.StatusConflict, "Request is still processing", "", "Retry with the same Idempotency-Key")
		}

		err = c.Next()
		if err != nil {
			if deleteErr := repo.Delete(c.UserContext(), record.ID); deleteErr != nil {
				return util.New(fiber.StatusInternalServerError, "IDEMPOTENCY_CLEANUP_FAILED", "Failed to clean up idempotency record", deleteErr.Error(), nil)
			}
			return err
		}

		body := c.Response().Body()
		if len(body) == 0 {
			body = []byte("null")
		}
		if err := repo.Complete(c.UserContext(), record.ID, c.Response().StatusCode(), body); err != nil {
			return util.New(fiber.StatusInternalServerError, "IDEMPOTENCY_COMPLETION_FAILED", "Failed to save idempotency response", err.Error(), nil)
		}
		return nil
	}
}