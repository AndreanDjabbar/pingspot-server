package router

import (
	authRouter "pingspot/internal/domain/auth_service/router"
	mainRouter "pingspot/internal/domain/report_service/router"
	searchRouter "pingspot/internal/domain/search_service/router"
	userRouter "pingspot/internal/domain/user_service/router"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	userRouter.RegisterUserRoutes(app)
	authRouter.RegisterAuthRoutes(app)
	searchRouter.RegisterSearchRoutes(app)
	mainRouter.RegisterReportRoutes(app)
}
