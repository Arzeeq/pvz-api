package server

import (
	"fmt"
	"net/http"

	"github.com/Arzeeq/pvz-api/internal/config"
	"github.com/Arzeeq/pvz-api/internal/dto"
	handler "github.com/Arzeeq/pvz-api/internal/handler/http"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type HTTPServer struct {
	cfg    *config.Config
	l      *logger.MyLogger
	router chi.Router
}

func NewHTTP(
	auth *handler.AuthHandler,
	pvz *handler.PVZHandler,
	reception *handler.ReceptionHandler,
	product *handler.ProductHandler,
	logger *logger.MyLogger,
	cfg *config.Config,
) (*HTTPServer, error) {
	r := chi.NewRouter()
	s := HTTPServer{
		cfg:    cfg,
		l:      logger,
		router: r,
	}

	// without authorization
	r.Post("/dummyLogin", auth.DummyLogin)
	r.Post("/register", auth.Register)
	r.Post("/login", auth.Login)

	// moderator only
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRoles(logger, []byte(cfg.JWTSecret), dto.UserRoleModerator))
		r.Post("/pvz", pvz.CreatePvz)
	})

	// employee only
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRoles(logger, []byte(cfg.JWTSecret), dto.UserRoleEmployee))
		r.Post("/receptions", reception.CreateReception)
		r.Post("/products", product.CreateProduct)
		r.Post("/pvz/{pvzId}/delete_last_product", pvz.DeleteLastProduct)
	})

	// moderator and employee
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthRoles(logger, []byte(cfg.JWTSecret), dto.UserRoleEmployee, dto.UserRoleModerator))
		r.Get("/pvz", pvz.GetPVZ)
		r.Post("/pvz/{pvzId}/close_last_reception", pvz.CloseReception)
	})

	return &s, nil
}

func (s *HTTPServer) Run() error {
	return http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", s.cfg.HTTPPort), s.router)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
