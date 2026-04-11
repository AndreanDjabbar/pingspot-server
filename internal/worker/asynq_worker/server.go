package asynqWorker

import (
	"fmt"
	"pingspot/internal/config"
	"pingspot/internal/worker/asynq_worker/handler"
	"pingspot/pkg/logger"

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

	opt := asynq.RedisClientOpt{
		Addr:     addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	return &WorkerServer{
		server: asynq.NewServer(
			opt,
			asynq.Config{
				Concurrency: 10,
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
