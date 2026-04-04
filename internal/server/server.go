package server

import (
	"pingspot/internal/middleware"
	"pingspot/internal/router"
	"pingspot/pkg/logger"
	env "pingspot/pkg/utils/env_util"
	response "pingspot/pkg/utils/response_util"
	_ "pingspot/docs"
	"github.com/gofiber/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
		AllowOrigins:     env.AllowedOrigins(),
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

	if env.NodeEnv() == "development" {
		defaultRoute.Get("/swagger/*", swagger.HandlerDefault)
	}

	router.RegisterRoutes(app)

	return &FiberServer{
		App: app,
	}
}

func DefaultHandler(c *fiber.Ctx) error {
	logger.Info("DEFAULT CONTROLLER")
	data := map[string]any{
		"message":    "Welcome to Pingspot API.. Please check the repository for more information.",
		"repository": env.GithubRepoURL(),
	}
	return response.ResponseSuccess(c, 200, "Welcome to Pingspot API", "data", data)
}