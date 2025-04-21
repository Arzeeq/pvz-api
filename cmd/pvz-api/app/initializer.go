package app

import (
	"errors"
	"time"

	"github.com/Arzeeq/pvz-api/internal/config"
	grpc_handler "github.com/Arzeeq/pvz-api/internal/handler/grpc"
	handler "github.com/Arzeeq/pvz-api/internal/handler/http"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/service"
	"github.com/Arzeeq/pvz-api/internal/storage/pg"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InitializeHandlers(pool *pgxpool.Pool, cfg *config.Config, logger *logger.MyLogger) (*Handlers, error) {
	if pool == nil || cfg == nil || logger == nil {
		return nil, errors.New("nil values in constructor")
	}

	storage, err := initStorages(pool)
	if err != nil {
		return nil, err
	}

	services, err := initServices(storage, cfg.JWTSecret, cfg.JWTDuration)
	if err != nil {
		return nil, err
	}

	handlers, err := initHandlers(services, logger, cfg.RequestTimeout)
	if err != nil {
		return nil, err
	}

	return handlers, nil
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

type Handlers struct {
	Auth      *handler.AuthHandler
	Product   *handler.ProductHandler
	Pvz       *handler.PVZHandler
	Reception *handler.ReceptionHandler
	GrpcPVZ   *grpc_handler.PVZHandler
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

func initHandlers(s *services, logger *logger.MyLogger, timeout time.Duration) (*Handlers, error) {
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
	return &Handlers{
		Auth:      authHandler,
		Product:   productHandler,
		Pvz:       pvzHandler,
		Reception: receptionHandler,
		GrpcPVZ:   grpcPvzHandler,
	}, nil
}
