package database

import (
	"fmt"
	"pingspot/internal/config"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"
	"sync"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

var (
	mongoDBInstance *mongo.Client
	mongoDBOnce     sync.Once
	mongoDBErr      error
)

func InitMongoDB(cfg config.MongoDBConfig) error {
	host := env.MongoHost()

	mongoDBOnce.Do(func() {
		var URI string
		if host != "localhost" {
			URI = fmt.Sprintf("mongodb://%s:%s@%s:%s/?ssl=true&replicaSet=globaldb&retrywrites=false&maxIdleTimeMS=120000&appName=@%s@",
                cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.User)
		} else {
			URI = "mongodb://" + cfg.User + ":" + cfg.Password + "@" + cfg.Host + ":" + cfg.Port

		}
		logger.Info("Connecting to MongoDB",
			zap.String("host", cfg.Host),
			zap.String("port", cfg.Port),
			zap.String("user", cfg.User),
		)
		var err error
		mongoDBInstance, err = mongo.Connect(options.Client().ApplyURI(URI))
		if err != nil {
			logger.Error("Failed to connect to MongoDB", zap.Error(err))
			mongoDBErr = fmt.Errorf("failed to connect to MongoDB: %w", err)
		} else {
			logger.Info("Connected to MongoDB successfully")
		}
	})
	return mongoDBErr
}

func GetMongoDB() *mongo.Client {
	return mongoDBInstance
}