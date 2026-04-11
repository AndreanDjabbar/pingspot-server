package config

import env "pingspot/pkg/utils/env_util"

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
		Host:     env.RedisHost(),
		Port:     env.RedisPort(),
		Username: env.RedisUsername(),
		Password: env.RedisPassword(),
		DB:       0,
		UseTLS:   env.RedisTLS(),
	}
}
