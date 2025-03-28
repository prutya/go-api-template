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
	if err := db.PingContext(ctx); err == nil {
		logger.Info("Database OK")
	} else {
		logger.Fatal("Failed to connect to the database", zap.Error(err))
	}

	// Repositories
	userRepo := repo.NewUserRepo(db)
	sessionRepo := repo.NewSessionRepo(db)

	// Services
	authenticationService := authentication_service.NewAuthenticationService(cfg, userRepo, sessionRepo)

	// Server
	router := server.NewRouter(cfg, logger, authenticationService)
	server := server.NewServer(cfg, router, logger)

	if err := server.Start(); err != nil {
		logger.Fatal("Server error", zap.Error(err))
	}

	logger.Info("Bye!")
}
