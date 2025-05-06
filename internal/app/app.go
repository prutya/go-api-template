package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/db"
	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/user_service"
	"prutya/go-api-template/internal/tasks_client"
)

type AppEssentials struct {
	Config  *config.Config
	Logger  *slog.Logger
	Context context.Context
}

type App struct {
	Essentials *AppEssentials

	DB bun.IDB

	TasksClient tasks_client.Client

	UserRepository         repo.UserRepo
	SessionRepository      repo.SessionRepo
	RefreshTokenRepository repo.RefreshTokenRepo
	AccessTokenRepository  repo.AccessTokenRepo

	AuthenticationService authentication_service.AuthenticationService
	UserService           user_service.UserService
}

func InitializeEssentials() *AppEssentials {
	// Config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Logger
	logger, err := loggerpkg.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		panic(err)
	}

	// Context
	ctx := context.Background()
	ctx = loggerpkg.NewContext(ctx, logger)

	return &AppEssentials{
		Config:  cfg,
		Logger:  logger,
		Context: ctx,
	}
}

func Initialize() *App {
	appEssentials := InitializeEssentials()

	cfg, logger, ctx := appEssentials.Config, appEssentials.Logger, appEssentials.Context

	// Database
	db := db.New(cfg.DatabaseUrl)

	// Smoke-test the database connection
	if err := db.PingContext(ctx); err == nil {
		logger.InfoContext(ctx, "Database OK")
	} else {
		logger.ErrorContext(ctx, "Failed to ping the database", "error", err)
		os.Exit(1)
	}

	// Tasks client
	tasksClient := tasks_client.NewClient(cfg.TasksRedisAddr, cfg.TasksRedisPassword)

	// Smoke-test the tasks client connection
	if err := tasksClient.Ping(); err != nil {
		logger.ErrorContext(ctx, "Failed to connect to the tasks client", "error", err)
	}
	logger.InfoContext(ctx, "Tasks client OK")

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

	return &App{
		Essentials:             appEssentials,
		DB:                     db,
		TasksClient:            tasksClient,
		UserRepository:         userRepo,
		SessionRepository:      sessionRepo,
		RefreshTokenRepository: refreshTokenRepo,
		AccessTokenRepository:  accessTokenRepo,
		AuthenticationService:  authenticationService,
		UserService:            userService,
	}
}
