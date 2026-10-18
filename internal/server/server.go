package server

import (
	"pingspot/internal/logger"
	"pingspot/internal/middleware"
	"pingspot/internal/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

type FiberServer struct {
	*fiber.App
}

func New() *FiberServer {
	app := fiber.New(fiber.Config{
		ServerHeader: "Pingspot Server",
		AppName:      "Pingspot API Server",
		BodyLimit:    10 * 1024 * 1024,
	})

	app.Static("/user", "./uploads/user")
	app.Static("/main", "./uploads/main")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     util.AllowedOrigins(),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type,X-Requested-With",
		ExposeHeaders:    "Set-Cookie",
		AllowCredentials: true,
		MaxAge:           300,
	}))

	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.LoggingMiddleware())
	app.Use(middleware.GlobalRateLimiterMiddleware())

	defaultRoute := app.Group("/pingspot/api")

	if util.NodeEnv() == "development" {
		defaultRoute.Get("/swagger/*", swagger.HandlerDefault)
	}

	RegisterRoutes(app)

	return &FiberServer{
		App: app,
	}
}

func DefaultHandler(c *fiber.Ctx) error {
	logger.Info("DEFAULT CONTROLLER")
	data := map[string]any{
		"message":    "Welcome to Pingspot API.. Please check the repository for more information.",
		"repository": util.GithubRepoURL(),
	}
	return util.ResponseSuccess(c, 200, "Welcome to Pingspot API", "data", data)
}