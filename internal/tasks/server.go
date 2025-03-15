package tasks

import (
	"context"
	"prutya/go-api-template/internal/config"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type Server struct {
	asynqServer *asynq.Server
	asynqMux    *asynq.ServeMux
}

func NewServer(ctx context.Context, config *config.Config, logger *zap.Logger) *Server {
	// TODO: Make server configurable via config fileE
	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: config.TasksRedisHost, Password: config.TasksRedisPassword},
		asynq.Config{
			Logger:      NewLoggerAdapter(logger.With(zap.String("type", "asynq"))),
			BaseContext: func() context.Context { return ctx },
			Concurrency: 10,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	asynqMux := asynq.NewServeMux()
	asynqMux.HandleFunc(TypeDemo, HandleDemoTask)

	server := &Server{
		asynqServer: asynqServer,
		asynqMux:    asynqMux,
	}

	return server
}

func (s *Server) Ping() error {
	return s.asynqServer.Ping()
}

func (s *Server) Run() error {
	return s.asynqServer.Run(s.asynqMux)
}
