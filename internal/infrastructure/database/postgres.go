package database

import (
	"fmt"
	"sync"

	"pingspot/internal/config"
	"pingspot/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	postgresDBInstance *gorm.DB
	postgresDBOnce     sync.Once
	postgresDBErr      error
)

func InitPostgres(cfg config.PostgresConfig) error {
	postgresDBOnce.Do(func() {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
		)

		logger.Info("Connecting to PostgreSQL",
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port),
			zap.String("user", cfg.User),
			zap.String("dbname", cfg.DBName),
		)

		var err error
		postgresDBInstance, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			logger.Error("Failed to connect to PostgreSQL", zap.Error(err))
			postgresDBErr = fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		} else {
			logger.Info("Connected to PostgreSQL successfully")
		}
	})
	return postgresDBErr
}

func GetPostgresDB() *gorm.DB {
	return postgresDBInstance
}
