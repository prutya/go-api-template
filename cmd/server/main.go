package main

import (
	"fmt"
	"prutya/todo/internal/logger"

	"prutya/todo/internal/config"
)

func main() {
	// Initialize the config
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Initialize the logger
	logger, err := logger.New(cfg.LogLevel, cfg.LogTimeFormat)
	if err != nil {
		panic(err)
	}
	logger.Debug("Logger OK")

	fmt.Println("Hello World!")
}
