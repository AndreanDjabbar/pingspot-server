package main

import (
	"fmt"
	"os"
	"pingspot/internal/config"
	"pingspot/internal/infrastructure/cache"
	"pingspot/internal/infrastructure/database"
	"pingspot/internal/migration"
	"pingspot/internal/server"
	"pingspot/internal/worker/asynqWorker"
	"pingspot/internal/worker/cronWorker"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"
	"strconv"

	"go.uber.org/zap"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	if err := config.LoadEnvConfig(); err != nil {
		panic(fmt.Sprintf("failed to load env config: %v", err))
	}

	err := logger.InitLogger("")
	defer logger.Logger.Sync()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	postgresConfig := config.LoadPostgresConfig()
	if err := database.InitPostgres(postgresConfig); err != nil {
		logger.Error("Failed to initialize PostgreSQL", zap.Error(err))
		panic(fmt.Sprintf("failed to initialize PostgreSQL: %v", err))
	}

	mongoConfig := config.LoadMongoDBConfig()
	if err := database.InitMongoDB(mongoConfig); err != nil {
		logger.Error("Failed to initialize MongoDB", zap.Error(err))
		panic(fmt.Sprintf("failed to initialize MongoDB: %v", err))
	}

	redisConfig := config.LoadRedisConfig()
	if err := cache.InitRedis(redisConfig); err != nil {
		logger.Error("Failed to initialize Redis", zap.Error(err))
		panic(fmt.Sprintf("failed to initialize Redis: %v", err))
	}

	if err := config.InitGoogleAuth(); err != nil {
		logger.Error("Failed to initialize Google Auth", zap.Error(err))
		panic(fmt.Sprintf("failed to initialize Google Auth: %v", err))
	}

	db := database.GetPostgresDB()
	if err := migration.Migrate(db); err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		panic(fmt.Sprintf("failed to run migrations: %v", err))
	}

	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		if err := os.MkdirAll("uploads/user", os.ModePerm); err != nil {
			logger.Error("Failed to create uploads/user directory", zap.Error(err))
		}

		if err := os.MkdirAll("uploads/main/report", os.ModePerm); err != nil {
			logger.Error("Failed to create uploads/main directory", zap.Error(err))
		}

		if err := os.MkdirAll("uploads/main/report/progress", os.ModePerm); err != nil {
			logger.Error("Failed to create uploads/main/report/progress directory", zap.Error(err))
		}
		
		if err := os.MkdirAll("uploads/main/report/comments", os.ModePerm); err != nil {
			logger.Error("Failed to create uploads/main/report/comments directory", zap.Error(err))
		}
	}

	workerServer := asynqWorker.NewWorkerServer(redisConfig)
	client := workerServer.GetClient()
	defer client.Close()
	
	go func() {
		logger.Info("Starting cron jobs")
		cronWorker.StartCron(client)
	}()

	go func() {
		logger.Info("Starting Asynq worker server")
		if err := workerServer.Run(); err != nil {
			logger.Error("Asynq worker server error", zap.Error(err))
		}
	}()

	server := server.New()

	server.RegisterFiberRoutes()

	port, _ := strconv.Atoi(env.Port())
	host := env.Host()
	logger.Info("Starting server on port", zap.Int("port", port), zap.String("host", host))
	err = server.Listen(fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		logger.Error("Failed to start server", zap.Error(err))
		panic(fmt.Sprintf("http server error: %s", err))
	}

	logger.Info("Server stopped.")
}
