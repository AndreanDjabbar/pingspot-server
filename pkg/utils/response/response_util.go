package response

import "github.com/gofiber/fiber/v2"

func ResponseSuccess(c *fiber.Ctx, status int, message string, key string, data any) error {
	if key == "" {
		key = "data"
	}
	response := fiber.Map{
		"success": true,
		"message": message,
		key:      data,
	}
	return c.Status(status).JSON(response)
}

func ResponseError(c *fiber.Ctx, status int, message string, key string, data any) error {
	if key == "" {
		key = "errors"
	}
	response := fiber.Map{
		"success": false,
		"message": message,
		key:      data,
	}
	return c.Status(status).JSON(response)
}