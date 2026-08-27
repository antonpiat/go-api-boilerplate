package main

import (
	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/antonpiat/go-api-boilerplate/internal/logger"
)

func main() {
	config, err := config.LoadConfig("config.yaml")
	if err != nil {
		panic(err)
	}

	logger := logger.NewLogger(config.LoggingConfig)
	logger.Info("Starting server")

	//TODO: database connection
	//TODO: middleware
	//TODO: authentication
	//TODO: authorization
	//TODO: router
	//TODO: metrics
	//TODO: tracing
	//TODO: security
	//TODO: start server
}
