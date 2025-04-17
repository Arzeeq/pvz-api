package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Arzeeq/pvz-api/internal/config"
	handler "github.com/Arzeeq/pvz-api/internal/handler/http"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/service"
	"github.com/Arzeeq/pvz-api/internal/storage/pg"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	cfg  *config.Config
	l    *logger.MyLogger
	r    chi.Router
	pool *pgxpool.Pool
}

func New(cfg *config.Config, logger *logger.MyLogger) (*Server, error) {
	if cfg == nil || logger == nil {
		return nil, errors.New("cfg and logger must be non nil")
	}

	s := Server{cfg: cfg, l: logger}
	pool, err := pg.InitDB(cfg.DBParam.GetConnStr())
	if err != nil {
		return nil, err
	}
	s.pool = pool

	// creating storages
	userStorage := pg.NewUserStorage(pool)

	// creating services
	tokenService := service.NewTokenService(cfg.JWTSecret, cfg.JWTDuration)
	userService, err := service.NewUserService(userStorage, tokenService)
	if err != nil {
		return nil, err
	}

	// creting handlers
	var authHandler *handler.Auth
	if authHandler, err = handler.NewAuthHandler(userService, tokenService, logger, cfg.RequestTimeout); err != nil {
		return nil, err
	}

	// mounting router
	r := chi.NewRouter()
	r.Post("/dummyLogin", authHandler.DummyLogin)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)
	s.r = r

	return &s, nil
}

func (s *Server) Run() error {
	s.l.Info("Starting application")

	migrator := pg.NewMigrator(s.cfg.MigrationDir, s.cfg.DBParam.GetConnStr())
	if err := migrator.Up(); err != nil {
		return err
	}

	s.l.Info("Application has started", slog.Int("port", s.cfg.ServerPort))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.ServerPort), s.r)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.pool.Close()
	return nil
}
