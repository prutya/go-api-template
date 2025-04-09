package tasks_server

import (
	"context"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/services/user_service"
	"prutya/go-api-template/internal/tasks"
)

type Server interface {
	Run() error
}

type server struct {
	asynqServer *asynq.Server
	asynqMux    *asynq.ServeMux
}

// TODO: Support more configuration parameters
func NewServer(
	baseCtx context.Context,
	redisAddr string,
	redisPassword string,
	userService user_service.UserService,
) Server {
	logger := loggerpkg.MustFromContext(baseCtx)

	srv := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr:     redisAddr,
			Password: redisPassword,
		},
		asynq.Config{
			Concurrency: 10,
			BaseContext: func() context.Context { return baseCtx },
			Logger:      newZapLoggerAdapter(logger),
		},
	)

	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware)
	mux.Handle(tasks.TypeUserHello, newUserHelloTaskHandler(userService))

	return &server{
		asynqServer: srv,
		asynqMux:    mux,
	}
}

func (s *server) Run() error {
	return s.asynqServer.Run(s.asynqMux)
}

func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		logger := loggerpkg.MustFromContext(ctx)

		taskID := t.ResultWriter().TaskID()

		// Add task ID to the logger context for better traceability
		logger = logger.With(zap.String("task_id", taskID))
		ctx = loggerpkg.NewContext(ctx, logger)

		start := time.Now()
		logger.Info("Processing task", zap.String("task_type", t.Type()))

		err := h.ProcessTask(ctx, t)
		if err != nil {

			logger.Error("Failed to process task", zap.Error(err), zap.Duration("duration", time.Since(start)))

			return err
		}

		logger.Info("Task processed", zap.Duration("duration", time.Since(start)))

		return nil
	})
}
