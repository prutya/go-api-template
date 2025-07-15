package app

import (
	"context"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/config"
	"prutya/go-api-template/internal/db"
	loggerpkg "prutya/go-api-template/internal/logger"
	"prutya/go-api-template/internal/repo"
	"prutya/go-api-template/internal/services/authentication_service"
	"prutya/go-api-template/internal/services/captcha_service"
	"prutya/go-api-template/internal/services/transactional_email_service"
	"prutya/go-api-template/internal/services/user_service"
	"prutya/go-api-template/internal/tasks_client"
)

type AppEssentials struct {
	Config  *config.Config
	Logger  *loggerpkg.Logger
	Context context.Context
}

type App struct {
	Essentials *AppEssentials

	DB          bun.IDB
	RepoFactory repo.RepoFactory

	TasksClient tasks_client.Client

	TransactionalEmailService transactional_email_service.TransactionalEmailService
	CaptchaService            captcha_service.CaptchaService
	AuthenticationService     authentication_service.AuthenticationService
	UserService               user_service.UserService
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
	db, err := db.New(
		cfg.DatabaseUrl,
		cfg.DatabaseMaxOpenConns,
		cfg.DatabaseMaxIdleConns,
		cfg.DatabaseMaxConnLifetime,
		cfg.DatabaseMaxConnIdleTime,
	)
	if err != nil {
		logger.FatalContext(ctx, "Failed to connect to the database", "error", err)
	}

	// Smoke-test the database connection
	if err := db.PingContext(ctx); err == nil {
		logger.InfoContext(ctx, "Database OK")
	} else {
		logger.FatalContext(ctx, "Failed to ping the database", "error", err)
	}

	// Tasks client
	tasksClient := tasks_client.NewClient(cfg.TasksRedisAddr, cfg.TasksRedisPassword)
	if err := tasksClient.Ping(); err == nil {
		logger.InfoContext(ctx, "Tasks client OK")
	} else {
		logger.FatalContext(ctx, "Failed to ping tasks client", "error", err)
	}

	// Repositories factory
	repoFactory := repo.NewRepoFactory()

	// Services
	transactionalEmailService, err := transactional_email_service.NewTransactionalEmailService(
		cfg.TransactionalEmailsEnabled,
		cfg.TransactionalEmailsDailyGlobalLimit,
		cfg.TransactionalEmailsSenderEmail,
		cfg.TransactionalEmailsSenderName,
		cfg.TransactionalEmailsScalewayAccessKeyID,
		cfg.TransactionalEmailsScalewaySecretKey,
		cfg.TransactionalEmailsScalewayRegion,
		cfg.TransactionalEmailsScalewayProjectID,
		db,
		repoFactory,
	)
	if err != nil {
		logger.FatalContext(ctx, "Failed to create transactional email service", "error", err)
	}

	captchaService := captcha_service.NewCaptchaService(
		cfg.CaptchaEnabled,
		cfg.CaptchaTurnstileBaseURL,
		cfg.CaptchaTurnstileSecretKey,
	)

	authenticationService := authentication_service.NewAuthenticationService(
		cfg,
		db,
		repoFactory,
		tasksClient,
		transactionalEmailService,
	)
	userService := user_service.NewUserService(db, repoFactory)

	return &App{
		Essentials: appEssentials,

		DB: db,

		TasksClient: tasksClient,

		CaptchaService:            captchaService,
		TransactionalEmailService: transactionalEmailService,
		AuthenticationService:     authenticationService,
		UserService:               userService,
	}
}
