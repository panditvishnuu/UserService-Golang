package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/panditvishnuu/userservice/internal/config"
	"github.com/panditvishnuu/userservice/internal/handler"
	"github.com/panditvishnuu/userservice/internal/repository"
	"github.com/panditvishnuu/userservice/internal/repository/postgres"
	"github.com/panditvishnuu/userservice/internal/server"
	"github.com/panditvishnuu/userservice/internal/service"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("unable to load config", "error", err)
		os.Exit(1)
	}

	// open a Db connection
	db, err := sql.Open("postgres", cfg.DSN())
	if err != nil {

		slog.Error("unable to open DB connection", "error", err)
		os.Exit(1)
	}

	var repo repository.UserRepo
	for i := 0; i < cfg.DBPingAttempts; i++ {
		attemptCtx, attemptCancel := context.WithTimeout(context.Background(), time.Duration(cfg.DBPingTimeout)*time.Second)
		repo, err = postgres.NewWithPing(attemptCtx, db)
		attemptCancel() // cancel immediately after use, don't defer inside a loop
		if err == nil {
			slog.Info("database connected successfully", "attempt", i+1)
			break
		}

		slog.Warn("database ping failed",
			"attempt", i+1,
			"error", err,
		)

		// sleep before next attempt
		if i < cfg.DBPingAttempts-1 {
			time.Sleep(time.Duration(cfg.DBPingDelaySeconds) * time.Second)
		}
	}
	if repo == nil {
		slog.Error("failed to connect to DB after retries",
			"attempts", cfg.DBPingAttempts,
			"last_error", err,
		)
		os.Exit(1)
	}

	s, err := service.NewUserService(repo, cfg.JWTSecret, cfg.JWTExpiration)
	if err != nil {
		slog.Error("failed to initialize user service", "error", err)
		os.Exit(1)
	}

	userHandler := handler.NewUserHandler(s)
	srv := server.New(cfg, userHandler)

	slog.Info("application started successfully")
	if err := srv.Run(); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}
}
