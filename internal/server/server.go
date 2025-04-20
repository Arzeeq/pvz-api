package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/dto"
	handler "github.com/Arzeeq/pvz-api/internal/handler/http"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/middleware"
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
	var pvzStorage *pg.PVZStorage
	var userStorage *pg.UserStorage
	var receptionStorage *pg.ReceptionStorage
	var productStorage *pg.ProductStorage
	if userStorage, err = pg.NewUserStorage(pool); err != nil {
		return nil, err
	}
	if pvzStorage, err = pg.NewPVZStorage(pool); err != nil {
		return nil, err
	}
	if receptionStorage, err = pg.NewReceptionStorage(pool); err != nil {
		return nil, err
	}
	if productStorage, err = pg.NewProductStorage(pool); err != nil {
		return nil, err
	}

	// creating services
	var pvzService *service.PVZService
	var userService *service.UserService
	var receptionService *service.ReceptionService
	var productService *service.ProductService
	var tokenService *service.TokenService
	if pvzService, err = service.NewPVZService(pvzStorage, receptionStorage, productStorage); err != nil {
		return nil, err
	}
	if userService, err = service.NewUserService(userStorage, tokenService); err != nil {
		return nil, err
	}
	if receptionService, err = service.NewReceptionService(receptionStorage); err != nil {
		return nil, err
	}
	if productService, err = service.NewProductService(productStorage); err != nil {
		return nil, err
	}
	if tokenService, err = service.NewTokenService([]byte(cfg.JWTSecret), cfg.JWTDuration); err != nil {
		return nil, err
	}

	// creting handlers
	var authHandler *handler.AuthHandler
	var pvzHandler *handler.PVZHandler
	var receptionHandler *handler.ReceptionHandler
	var productHandler *handler.ProductHandler
	if authHandler, err = handler.NewAuthHandler(userService, tokenService, logger, cfg.RequestTimeout); err != nil {
		return nil, err
	}
	if pvzHandler, err = handler.NewPvzHandler(pvzService, receptionService, productService, logger, cfg.RequestTimeout); err != nil {
		return nil, err
	}
	if receptionHandler, err = handler.NewReceptionHandler(receptionService, logger, cfg.RequestTimeout); err != nil {
		return nil, err
	}
	if productHandler, err = handler.NewProductHandler(productService, logger, cfg.RequestTimeout); err != nil {
		return nil, err
	}

	// mounting router
	r := chi.NewRouter()
	s.r = r

	// without authorization
	r.Post("/dummyLogin", authHandler.DummyLogin)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	// moderator only
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRoles(logger, []byte(cfg.JWTSecret), dto.UserRoleModerator))

		r.Post("/pvz", pvzHandler.CreatePvz)
	})

	// employee only
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRoles(logger, []byte(cfg.JWTSecret), dto.UserRoleEmployee))

		r.Post("/receptions", receptionHandler.CreateReception)
		r.Post("/products", productHandler.CreateProduct)
		r.Post("/pvz/{pvzId}/delete_last_product", pvzHandler.DeleteLastProduct)
	})

	// moderator and employee
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRoles(logger, []byte(cfg.JWTSecret), dto.UserRoleEmployee, dto.UserRoleModerator))

		r.Get("/pvz", pvzHandler.GetPVZ)
		r.Post("/pvz/{pvzId}/close_last_reception", pvzHandler.CloseReception)
	})

	return &s, nil
}

func (s *Server) Run() error {
	s.l.Info("Starting application")

	migrator := pg.NewMigrator(s.cfg.MigrationDir, s.cfg.DBParam.GetConnStr())
	if err := migrator.Up(); err != nil {
		return err
	}
	s.l.Info("Migrations have been done")

	s.l.Info("Application has started", slog.Int("port", s.cfg.ServerPort))
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.ServerPort), s.r)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.pool.Close()
	return nil
}
