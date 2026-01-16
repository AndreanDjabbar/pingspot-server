package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func LoadEnvConfig() error {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	envFile := getEnvFile(env)

	if err := godotenv.Load(envFile); err != nil {
		return fmt.Errorf("failed to load env file %s: %w", envFile, err)
	}

	return nil
}

func getEnvFile(env string) string {
	switch env {
	case "development":
		return ".env.dev"
	case "production":
		return ".env.prod"
	case "testing":
		return ".env.test"
	case "staging":
		return ".env.staging"
	default:
		return ".env"
	}
}
