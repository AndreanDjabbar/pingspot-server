package asynqWorker

import (
	"crypto/tls"
	"fmt"
	"pingspot/internal/config"
	"pingspot/internal/worker/asynqWorker/handler"
	"pingspot/pkg/logger"
	"pingspot/pkg/utils/env"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type WorkerServer struct {
	server    *asynq.Server
	client    *asynq.Client
	redisAddr string
}

func NewWorkerServer(cfg config.RedisConfig) *WorkerServer {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	var opt asynq.RedisConnOpt

	if env.RedisHost() != "localhost" {
		opt = asynq.RedisClusterClientOpt{
			Addrs:    []string{addr},
            Username: cfg.Username,
            Password: cfg.Password,
            TLSConfig: &tls.Config{
                MinVersion:         tls.VersionTLS12,
                InsecureSkipVerify: true, 
            },
		}
	} else {
		opt = asynq.RedisClientOpt{
			Addr: addr,
			DB:   cfg.DB,
		}
	}

	return &WorkerServer{
		server: asynq.NewServer(
			opt,
			asynq.Config{
				Concurrency: 10,
				// Queues: map[string]int{
				// 	"default": 6,
				// 	"critical": 4,
				// },
			},
		),
		client:    asynq.NewClient(opt),
		redisAddr: addr,
	}
}

func (w *WorkerServer) Run() error {
	mux := asynq.NewServeMux()

	handler.RegisterAllHandlers(mux)
	if err := w.server.Run(mux); err != nil {
		logger.Error("❌ Asynq server failed to start", zap.Error(err))
		return err
	}

	return nil
}

func (w *WorkerServer) GetClient() *asynq.Client {
	return w.client
}

func (w *WorkerServer) Stop() {
	w.server.Stop()
	w.server.Shutdown()
}
