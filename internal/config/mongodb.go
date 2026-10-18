package config

import "pingspot/internal/util"

type MongoDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

func LoadMongoDBConfig() MongoDBConfig {
	return MongoDBConfig{
		Host:     util.MongoHost(),
		Port:     util.MongoPort(),
		User:     util.MongoUser(),
		Password: util.MongoPassword(),
	}
}
