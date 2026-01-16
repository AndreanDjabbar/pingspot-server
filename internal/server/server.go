package server

import (
	"pingspot/internal/middleware"
	"pingspot/internal/router"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"
	"pingspot/pkg/utils/response"

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

	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.LoggingMiddleware())
	app.Use(middleware.GlobalRateLimiterMiddleware())

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

func (s *FiberServer) RegisterFiberRoutes() {
	env := env.NodeEnv()
	var origin string
	if env != "production" {
		origin = "http://localhost:3000"
	} else {
		origin = "https://pingspot.vercel.app"
	}

	s.App.Use(cors.New(cors.Config{
		AllowOrigins:     origin,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Accept,Authorization,Content-Type",
		AllowCredentials: true,
		MaxAge:           300,
	}))
	defaultRoute := s.App.Group("/pingspot/api")
	defaultRoute.Get("/", DefaultHandler)

	router.RegisterRoutes(s.App)
}
