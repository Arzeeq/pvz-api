package app

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Arzeeq/pvz-api/internal/config"
	grpc_handler "github.com/Arzeeq/pvz-api/internal/handler/grpc"
	handler "github.com/Arzeeq/pvz-api/internal/handler/http"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/server"
	"github.com/Arzeeq/pvz-api/internal/service"
	"github.com/Arzeeq/pvz-api/internal/storage/pg"
	"github.com/jackc/pgx/v5/pgxpool"
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

	storages, err := initStorages(pool)
	if err != nil {
		return nil, deferFn, err
	}

	services, err := initServices(storages, cfg.JWTSecret, cfg.JWTDuration)
	if err != nil {
		return nil, deferFn, err
	}

	handlers, err := initHandlers(services, logger, cfg.RequestTimeout)
	if err != nil {
		return nil, deferFn, err
	}

	http, err := server.NewHTTP(handlers.auth, handlers.pvz, handlers.reception, handlers.product, logger, cfg)
	if err != nil {
		return nil, deferFn, err
	}

	grpc, err := server.NewGRPC(handlers.grpcPVZ)
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
	return nil
}

type storages struct {
	product   *pg.ProductStorage
	pvz       *pg.PVZStorage
	reception *pg.ReceptionStorage
	user      *pg.UserStorage
}

type services struct {
	product   *service.ProductService
	pvz       *service.PVZService
	reception *service.ReceptionService
	token     *service.TokenService
	user      *service.UserService
}

type handlers struct {
	auth      *handler.AuthHandler
	product   *handler.ProductHandler
	pvz       *handler.PVZHandler
	reception *handler.ReceptionHandler
	grpcPVZ   *grpc_handler.PVZHandler
}

func initStorages(pool *pgxpool.Pool) (*storages, error) {
	var productStorage *pg.ProductStorage
	var pvzStorage *pg.PVZStorage
	var receptionStorage *pg.ReceptionStorage
	var userStorage *pg.UserStorage
	var err error
	if productStorage, err = pg.NewProductStorage(pool); err != nil {
		return nil, err
	}
	if pvzStorage, err = pg.NewPVZStorage(pool); err != nil {
		return nil, err
	}
	if receptionStorage, err = pg.NewReceptionStorage(pool); err != nil {
		return nil, err
	}
	if userStorage, err = pg.NewUserStorage(pool); err != nil {
		return nil, err
	}
	return &storages{
		product:   productStorage,
		pvz:       pvzStorage,
		reception: receptionStorage,
		user:      userStorage,
	}, nil
}

func initServices(storage *storages, jwtSecret string, jwtDuration time.Duration) (*services, error) {
	var productService *service.ProductService
	var pvzService *service.PVZService
	var receptionService *service.ReceptionService
	var tokenService *service.TokenService
	var userService *service.UserService
	var err error
	if productService, err = service.NewProductService(storage.product); err != nil {
		return nil, err
	}
	if pvzService, err = service.NewPVZService(storage.pvz, storage.reception, storage.product); err != nil {
		return nil, err
	}
	if receptionService, err = service.NewReceptionService(storage.reception); err != nil {
		return nil, err
	}
	if tokenService, err = service.NewTokenService([]byte(jwtSecret), jwtDuration); err != nil {
		return nil, err
	}
	if userService, err = service.NewUserService(storage.user, tokenService); err != nil {
		return nil, err
	}
	return &services{
		product:   productService,
		pvz:       pvzService,
		reception: receptionService,
		token:     tokenService,
		user:      userService,
	}, nil
}

func initHandlers(s *services, logger *logger.MyLogger, timeout time.Duration) (*handlers, error) {
	var authHandler *handler.AuthHandler
	var productHandler *handler.ProductHandler
	var pvzHandler *handler.PVZHandler
	var receptionHandler *handler.ReceptionHandler
	var grpcPvzHandler *grpc_handler.PVZHandler
	var err error
	if authHandler, err = handler.NewAuthHandler(s.user, s.token, logger, timeout); err != nil {
		return nil, err
	}
	if productHandler, err = handler.NewProductHandler(s.product, logger, timeout); err != nil {
		return nil, err
	}
	if pvzHandler, err = handler.NewPvzHandler(s.pvz, s.reception, s.product, logger, timeout); err != nil {
		return nil, err
	}
	if receptionHandler, err = handler.NewReceptionHandler(s.reception, logger, timeout); err != nil {
		return nil, err
	}
	if grpcPvzHandler, err = grpc_handler.NewPVZHandler(s.pvz); err != nil {
		return nil, err
	}
	return &handlers{
		auth:      authHandler,
		product:   productHandler,
		pvz:       pvzHandler,
		reception: receptionHandler,
		grpcPVZ:   grpcPvzHandler,
	}, nil
}
