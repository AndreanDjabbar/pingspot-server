package config

import "pingspot/internal/util"

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     util.PostgreHost(),
		Port:     util.PostgrePort(),
		User:     util.PostgreUser(),
		Password: util.PostgrePassword(),
		DBName:   util.PostgreDB(),
		SSLMode:  util.PostgreSSLMode(),
	}
}
