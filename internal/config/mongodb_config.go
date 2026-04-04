package config

import env "pingspot/pkg/utils/env_util"

type MongoDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

func LoadMongoDBConfig() MongoDBConfig {
	return MongoDBConfig{
		Host:     env.MongoHost(),
		Port:     env.MongoPort(),
		User:     env.MongoUser(),
		Password: env.MongoPassword(),
	}
}
