package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antonpiat/go-api-boilerplate/internal/auth"
	"github.com/antonpiat/go-api-boilerplate/internal/cache"
	"github.com/antonpiat/go-api-boilerplate/internal/config"
	"github.com/antonpiat/go-api-boilerplate/internal/database"
	"github.com/antonpiat/go-api-boilerplate/internal/logger"
	"github.com/antonpiat/go-api-boilerplate/internal/server"
	"github.com/antonpiat/go-api-boilerplate/internal/user"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.Logging, cfg.App.Environment)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	log.Info("starting " + cfg.App.Name + " (" + cfg.App.Environment + ") on " + cfg.Server.Addr())

	connectCtx, connectCancel := context.WithTimeout(context.Background(), 10*time.Second)
	db, err := database.New(connectCtx, cfg.Database)
	if err != nil {
		connectCancel()
		log.Error("database connection failed", zap.Error(err))
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Error("close database", zap.Error(err))
		}
	}()

	rdb, err := cache.New(connectCtx, cfg.Redis)
	connectCancel()
	if err != nil {
		log.Error("redis connection failed", zap.Error(err))
		return err
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Error("close redis", zap.Error(err))
		}
	}()

	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	tokenSvc := auth.NewTokenService(cfg.JWT, cfg.App.Name)
	tokenStore := auth.NewRedisTokenStore(rdb)
	authSvc := auth.NewService(userRepo, tokenSvc, tokenStore)
	authHandler := auth.NewHandler(authSvc)

	engine := server.New(server.Dependencies{
		Config: cfg,
		Logger: log,
		Ready: server.CompositeReady{Checks: []func(context.Context) error{
			func(ctx context.Context) error { return database.Ping(ctx, db) },
			func(ctx context.Context) error { return cache.Ping(ctx, rdb) },
		}},
		AuthHandler: authHandler,
		AuthService: authSvc,
		UserHandler: userHandler,
	})

	httpServer := server.NewHTTPServer(cfg, engine)

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening on " + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-stop:
		log.Info("shutdown signal: " + sig.String())
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeoutDuration())
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", zap.Error(err))
		return err
	}
	log.Info("server stopped")
	return nil
}
