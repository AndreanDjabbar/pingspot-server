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
		fmt.Printf("Warning: could not load env file %s: %v\n", envFile, err)
		return nil
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
