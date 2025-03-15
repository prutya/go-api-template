package main

import (
	"context"

	"go.uber.org/zap"

	"prutya/go-api-template/internal/config"
	db "prutya/go-api-template/internal/db"
	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/tasks"
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

	// Background tasks server
	server := tasks.NewServer(ctx, cfg, logger)

	// Smoke-test the background tasks server
	if err := server.Ping(); err == nil {
		logger.Info("Background tasks server OK")
	} else {
		logger.Fatal("Failed to connect to the background tasks server", zap.Error(err))
	}

	// Run the background tasks server
	if err := server.Run(); err != nil {
		logger.Fatal("Failed to run the background tasks server", zap.Error(err))
	}
}
