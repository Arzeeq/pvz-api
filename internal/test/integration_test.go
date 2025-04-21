package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/dto"
	handler "github.com/Arzeeq/pvz-api/internal/handler/http"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/server"
	"github.com/Arzeeq/pvz-api/internal/service"
	"github.com/Arzeeq/pvz-api/internal/storage/pg"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var cfg = &config.Config{
	DBParam: config.DBParam{
		DBUser:     "postgres",
		DBPassword: "postgresPassword",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "test",
	},
	Env:            config.EnvTest,
	JWTDuration:    time.Hour,
	LoggerFormat:   logger.LogFormatText,
	MigrationDir:   "../../internal/storage/pg/migrations",
	RequestTimeout: 5 * time.Second,
	JWTSecret:      "MyJWTSecret",
	HTTPPort:       8080,
}

func TestIntegrationWithTestContainers(t *testing.T) {
	deferFn, err := createContainer(context.Background())
	require.NoError(t, err)
	defer deferFn()

	log := logger.New(cfg.Env, cfg.LoggerFormat)

	pool, closeConn, err := pg.InitDB(cfg.ConnectionStr)
	require.NoError(t, err)
	defer closeConn()

	storages, err := initStorages(pool)
	require.NoError(t, err)

	services, err := initServices(storages, cfg.JWTSecret, cfg.JWTDuration)
	require.NoError(t, err)

	handlers, err := initHandlers(services, log, cfg.RequestTimeout)
	require.NoError(t, err)

	server, err := server.NewHTTP(handlers.auth, handlers.pvz, handlers.reception, handlers.product, log, cfg)
	require.NoError(t, err, "Failed to create server")

	t.Run("create pvz, create reception, add 50 products, close reception", func(t *testing.T) {
		// get tokens
		tokenEmployee, err := getToken(server, dto.PostDummyLoginJSONBodyRoleEmployee)
		require.NoError(t, err)
		tokenModerator, err := getToken(server, dto.PostDummyLoginJSONBodyRoleModerator)
		require.NoError(t, err)

		// create pvz
		pvzRequest := &dto.PostPvzJSONRequestBody{City: dto.SaintPetersburg}
		req, err := newRequest("POST", "http://localhost:8080/pvz", tokenModerator, pvzRequest)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
		require.NotNil(t, w.Body)

		var pvzResponse dto.PVZ
		require.NoError(t, unmarshalResponse(w, &pvzResponse))
		require.NotNil(t, pvzResponse.Id)
		require.NotNil(t, pvzResponse.RegistrationDate)

		// create reception
		receptionRequest := &dto.PostReceptionsJSONBody{PvzId: *pvzResponse.Id}
		req, err = newRequest("POST", "http://localhost:8080/receptions", tokenEmployee, receptionRequest)
		require.NoError(t, err)

		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
		require.NotNil(t, w.Body)

		var receptionResponse dto.Reception
		require.NoError(t, unmarshalResponse(w, &receptionResponse))
		require.Equal(t, receptionRequest.PvzId, receptionResponse.PvzId)
		require.Equal(t, dto.InProgress, receptionResponse.Status)

		// add 50 products
		for i := range 50 {
			productRequest := &dto.PostProductsJSONBody{PvzId: *pvzResponse.Id}
			switch i % 3 {
			case 0:
				productRequest.Type = dto.PostProductsJSONBodyTypeElectronics
			case 1:
				productRequest.Type = dto.PostProductsJSONBodyTypeClothes
			case 2:
				productRequest.Type = dto.PostProductsJSONBodyTypeShoes
			}
			req, err = newRequest("POST", "http://localhost:8080/products", tokenEmployee, productRequest)
			require.NoError(t, err)

			w = httptest.NewRecorder()
			server.ServeHTTP(w, req)
			require.Equal(t, http.StatusCreated, w.Code)
			require.NotNil(t, w.Body)

			var productResponse dto.Product
			require.NoError(t, unmarshalResponse(w, &productResponse))
			require.NotNil(t, productResponse.Id)
			require.NotNil(t, productResponse.DateTime)
			require.Equal(t, receptionResponse.Id, productResponse.ReceptionId)
			require.Equal(t, string(productRequest.Type), string(productResponse.Type))
		}

		// close reception
		path := fmt.Sprintf("http://localhost:8080/pvz/%s/close_last_reception", pvzResponse.Id)
		req, err = newRequest("POST", path, tokenModerator, nil)
		require.NoError(t, err)

		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.NotNil(t, w.Body)

		var closeResponse dto.Reception
		require.NoError(t, unmarshalResponse(w, &closeResponse))
		require.Equal(t, receptionResponse.Id, closeResponse.Id)
		require.Equal(t, receptionResponse.DateTime, closeResponse.DateTime)
		require.Equal(t, receptionResponse.PvzId, closeResponse.PvzId)
		require.Equal(t, dto.Close, closeResponse.Status)
	})
}

func createContainer(ctx context.Context) (func(), error) {
	container, err := postgres.Run(ctx,
		"postgres:16.8",
		postgres.WithDatabase(cfg.DBName),
		postgres.WithUsername(cfg.DBUser),
		postgres.WithPassword(cfg.DBPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)))

	if err != nil {
		return nil, err
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}
	cfg.ConnectionStr = connStr

	terminate := func() {
		if err := container.Terminate(ctx); err != nil {
			log.Fatal("failed to terminate container")
		}
	}

	migrator := pg.NewMigrator(cfg.MigrationDir, connStr)
	if err := migrator.Up(); err != nil {
		terminate()
		return nil, err
	}

	return terminate, nil
}

func getToken(server *server.HTTPServer, role dto.PostDummyLoginJSONBodyRole) (string, error) {
	payload := &dto.PostDummyLoginJSONBody{Role: role}
	req, err := newRequest("POST", "http://localhost:8080/dummyLogin", "", payload)
	if err != nil {
		return "", nil
	}

	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		return "", errors.New("wrong status code in getToken")
	}
	if w.Body == nil {
		return "", errors.New("token was not provided")
	}

	var token dto.Token
	err = unmarshalResponse(w, &token)
	if err != nil {
		return "", err
	}

	return token, nil
}

func newRequest(method string, path string, token string, payload interface{}) (*http.Request, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))

	return req, nil
}

func unmarshalResponse(w *httptest.ResponseRecorder, payload interface{}) error {
	body := w.Body.Bytes()
	return json.Unmarshal(body, payload)
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
	return &handlers{
		auth:      authHandler,
		product:   productHandler,
		pvz:       pvzHandler,
		reception: receptionHandler,
	}, nil
}
