package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Arzeeq/pvz-api/cmd/pvz-api/app"
	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/logger"
)

func main() {
	// load config and create logger
	cfg := config.MustLoad(os.Getenv("CONFIG_PATH"))
	l := logger.New(cfg.Env, cfg.LoggerFormat)
	l.Info("config loaded successfully", slog.String("env", cfg.Env))

	app, deferFn, err := app.NewApplication(cfg, l)
	if err != nil {
		l.WrapError("failed to create application instance", err)
	}
	defer deferFn()

	go func() {
		if err := app.Run(); err != nil {
			l.WrapError("application has encountered an error", err)
		}
	}()

	// catching os signals SIGINT and SIGTERM
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	l.Info("Gracefully shutting down application")
}
