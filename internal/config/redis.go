package config

import "pingspot/internal/util"

type RedisConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DB       int
	UseTLS   bool
}

func LoadRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     util.RedisHost(),
		Port:     util.RedisPort(),
		Username: util.RedisUsername(),
		Password: util.RedisPassword(),
		DB:       0,
		UseTLS:   util.RedisTLS(),
	}
}
