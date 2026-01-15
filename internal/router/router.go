package router

import (
	authRouter "pingspot/internal/domain/authService/router"
	mainRouter "pingspot/internal/domain/reportService/router"
	searchRouter "pingspot/internal/domain/searchService/router"
	userRouter "pingspot/internal/domain/userService/router"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	userRouter.RegisterUserRoutes(app)
	authRouter.RegisterAuthRoutes(app)
	searchRouter.RegisterSearchRoutes(app)
	mainRouter.RegisterReportRoutes(app)
}
