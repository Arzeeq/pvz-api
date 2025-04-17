package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/server"
)

func main() {
	// load config and create logger
	cfg := config.MustLoad(os.Getenv("CONFIG_PATH"))
	l := logger.New(cfg.Env, cfg.LoggerFormat)
	l.Info("config loaded successfully", slog.String("env", cfg.Env))

	// create server instance
	s, err := server.New(cfg, l)
	if err != nil {
		l.WrapError("failed to create a server instance", err)
	}

	// run server
	go func() {
		if err := s.Run(); err != nil && err != http.ErrServerClosed {
			l.WrapError("application has encountered an error", err)
		}
	}()

	// catching os signals SIGINT and SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Gracefully shutting down server")

	// shutting down server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		l.WrapError("Server shutdown has encountered an error", err)
	}

	// waiting for timeout
	<-ctx.Done()
	l.Info("Server was shut down")
}
