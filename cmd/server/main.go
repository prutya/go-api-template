// TODO: Extract code that is shared with the worker into a common package
// "app container"?

package main

import (
	"context"

	"go.uber.org/zap"

	"prutya/go-api-template/internal/config"
	db "prutya/go-api-template/internal/db"
	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/server"
	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/user_service"
	"prutya/go-api-template/internal/tasks_client"
)

func main() {
	// Config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Logger
	logger, err := loggerpkg.New(cfg.LogLevel, cfg.LogTimeFormat)
	if err != nil {
		panic(err)
	}
	logger.Info("Logger OK")

	// Context
	ctx := context.Background()
	ctx = loggerpkg.NewContext(ctx, logger)

	// Database
	db := db.New(cfg.DatabaseUrl)

	// Smoke-test the database connection
	if err := db.PingContext(ctx); err != nil {
		logger.Fatal("Failed to connect to the database", zap.Error(err))
	}
	logger.Info("Database OK")

	// Tasks client
	tasksClient := tasks_client.NewClient(cfg.TasksRedisAddr, cfg.TasksRedisPassword)

	// Smoke-test the tasks client connection
	if err := tasksClient.Ping(); err != nil {
		logger.Fatal("Failed to connect to the tasks client", zap.Error(err))
	}
	logger.Info("Tasks client OK")

	// Repositories
	userRepo := repo.NewUserRepo(db)
	sessionRepo := repo.NewSessionRepo(db)
	refreshTokenRepo := repo.NewRefreshTokenRepo(db)
	accessTokenRepo := repo.NewAccessTokenRepo(db)

	// Services
	authenticationService := authentication_service.NewAuthenticationService(
		cfg,
		userRepo,
		sessionRepo,
		refreshTokenRepo,
		accessTokenRepo,
		tasksClient,
	)
	userService := user_service.NewUserService(userRepo)

	// Server
	router := server.NewRouter(cfg, logger, authenticationService, userService)
	server := server.NewServer(cfg, router, logger)

	if err := server.Start(); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}

	logger.Info("Bye!")
}
