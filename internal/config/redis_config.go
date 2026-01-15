package config

import "pingspot/pkg/utils/env"

type RedisConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
}

func LoadRedisConfig() RedisConfig {
	if env.RedisHost() != "localhost" {
		return RedisConfig{
			Host:     env.RedisHost(),
			Port:     env.RedisPort(),
			Username: env.RedisUsername(),
			Password: env.RedisPassword(),
			DB:       0,
		}
	}
	return RedisConfig{
		Host: env.RedisHost(),
		Port: env.RedisPort(),
		DB:   0,
	}
}
