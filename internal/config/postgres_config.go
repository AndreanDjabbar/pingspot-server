package config

import "pingspot/pkg/utils/env"

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadPostgresConfig() PostgresConfig {
	isRequiedSSL := env.PostgreHost() != "localhost"

	return PostgresConfig{
		Host:     env.PostgreHost(),
		Port:     env.PostgrePort(),
		User:     env.PostgreUser(),
		Password: env.PostgrePassword(),
		DBName:   env.PostgreDB(),
		SSLMode:  func() string {
			if isRequiedSSL {
				return "require"
			}
			return "disable"
		}(),
	}
}
