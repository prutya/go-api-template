package app

import (
	"context"

	"github.com/uptrace/bun"
	"go.uber.org/zap"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/db"
	"prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/user_service"
	"prutya/go-api-template/internal/tasks_client"
)

type App struct {
	Config  *config.Config
	Logger  *zap.Logger
	Context context.Context
	DB      bun.IDB

	TasksClient tasks_client.Client

	UserRepository         repo.UserRepo
	SessionRepository      repo.SessionRepo
	RefreshTokenRepository repo.RefreshTokenRepo
	AccessTokenRepository  repo.AccessTokenRepo

	AuthenticationService authentication_service.AuthenticationService
	UserService           user_service.UserService
}

func Initialize() *App {
	// Config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Logger
	loggerInstance, err := logger.New(cfg.LogLevel, cfg.LogTimeFormat)
	if err != nil {
		panic(err)
	}
	loggerInstance.Info("Logger OK")

	// Context
	ctx := context.Background()
	ctx = logger.NewContext(ctx, loggerInstance)

	// Database
	db := db.New(cfg.DatabaseUrl)

	// Smoke-test the database connection
	if err := db.PingContext(ctx); err != nil {
		loggerInstance.Fatal("Failed to connect to the database", zap.Error(err))
	}
	loggerInstance.Info("Database OK")

	// Tasks client
	tasksClient := tasks_client.NewClient(cfg.TasksRedisAddr, cfg.TasksRedisPassword)

	// Smoke-test the tasks client connection
	if err := tasksClient.Ping(); err != nil {
		loggerInstance.Fatal("Failed to connect to the tasks client", zap.Error(err))
	}
	loggerInstance.Info("Tasks client OK")

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
		Config:                 cfg,
		Logger:                 loggerInstance,
		Context:                ctx,
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
