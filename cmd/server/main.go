package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"university_agency/internal/auth"
	"university_agency/internal/config"
	"university_agency/internal/dashboard"
	"university_agency/internal/platform/database"
	"university_agency/internal/web"
)

func main() {
	cfg, err := config.Load(".env")
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: cfg.App.Env == "development",
		Level:     slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	db, err := database.Open(cfg.Database)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	repository, err := dashboard.NewSQLRepository(context.Background(), db, cfg.Database.Driver, dashboard.SQLRepositoryOptions{
		SeedDemoData: cfg.App.Env == "development",
	})
	if err != nil {
		logger.Error("initialize sql repository", "error", err)
		os.Exit(1)
	}
	logger.Info("database enabled; using sql repository", "driver", cfg.Database.Driver)

	dashboardService := dashboard.NewService(repository)
	sessionStore := auth.NewSessionStore(12 * time.Hour)

	server := &http.Server{
		Addr:              cfg.HTTP.Addr(),
		Handler:           web.NewRouter(web.Dependencies{Config: cfg, Logger: logger, Dashboard: dashboardService, Store: repository, Sessions: sessionStore}),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
	}

	errs := make(chan error, 1)
	go func() {
		logger.Info("server started", "addr", "http://"+cfg.HTTP.Addr())
		errs <- server.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-stop:
		logger.Info("shutdown requested", "signal", sig.String())
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("force close failed", "error", closeErr)
		}
		os.Exit(1)
	}

	logger.Info("server stopped")
}
