package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/server"
	"github.com/Arzeeq/pvz-api/internal/storage/pg"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

type Application struct {
	cfg  *config.Config
	l    *logger.MyLogger
	http *server.HTTPServer
	grpc *grpc.Server
}

func NewApplication(cfg *config.Config, logger *logger.MyLogger) (*Application, func(), error) {
	if cfg == nil || logger == nil {
		return nil, nil, errors.New("cfg and logger must be non nil")
	}

	pool, deferFn, err := pg.InitDB(cfg.ConnectionStr)
	if err != nil {
		return nil, nil, err
	}

	migrator := pg.NewMigrator(cfg.MigrationDir, cfg.ConnectionStr)
	if err := migrator.Up(); err != nil {
		return nil, deferFn, err
	}

	handlers, err := InitializeHandlers(pool, cfg, logger)
	if err != nil {
		return nil, deferFn, err
	}

	http, err := server.NewHTTP(handlers.Auth, handlers.Pvz, handlers.Reception, handlers.Product, logger, cfg)
	if err != nil {
		return nil, deferFn, err
	}

	grpc, err := server.NewGRPC(handlers.GrpcPVZ)
	if err != nil {
		return nil, deferFn, err
	}

	app := Application{
		cfg:  cfg,
		l:    logger,
		http: http,
		grpc: grpc,
	}

	return &app, deferFn, nil
}

func (app *Application) Run() error {
	app.l.Info("Running application")

	app.l.Info("Starting HTTP", slog.Int("port", app.cfg.HTTPPort))
	go func() {
		if err := app.http.Run(); err != nil {
			app.l.WrapError("HTTP server has encountered an error", err)
		}
	}()

	app.l.Info("Starting gRPC", slog.Int("port", app.cfg.GRPCPort))
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", app.cfg.GRPCPort))
		if err != nil {
			app.l.WrapError("failed to listen grpc port", err)
			return
		}

		if err := app.grpc.Serve(lis); err != nil {
			app.l.WrapError("failed to serve grpc port", err)
		}
	}()

	r := chi.NewRouter()
	r.Handle("/metrics", promhttp.Handler())
	app.l.Info("Starting Prometheus", slog.Int("port", app.cfg.PrometheusPort))
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", app.cfg.PrometheusPort), r); err != nil {
			app.l.WrapError("failed to serve prometheus port", err)
		}
	}()

	return nil
}
