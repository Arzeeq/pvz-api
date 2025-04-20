package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/go-playground/validator/v10"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type ProductServicer interface {
	CreateProduct(ctx context.Context, productDto dto.PostProductsJSONBody) (*dto.Product, error)
	DeleteLastProduct(ctx context.Context, pvzID openapi_types.UUID) error
}

type ProductHandler struct {
	productService ProductServicer
	log            *logger.MyLogger
	validator      *validator.Validate
	timeout        time.Duration
}

func NewProductHandler(
	productService ProductServicer,
	logger *logger.MyLogger,
	timeout time.Duration,
) (*ProductHandler, error) {
	if productService == nil || logger == nil {
		return nil, errors.New("nil values in NewProductHandler constructor")
	}

	return &ProductHandler{
		productService: productService,
		log:            logger,
		validator:      validator.New(),
		timeout:        timeout,
	}, nil
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var productDto dto.PostProductsJSONBody
	if err := dto.Parse(r.Body, &productDto); err != nil {
		h.log.HTTPError(w, http.StatusUnauthorized, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	product, err := h.productService.CreateProduct(ctx, productDto)
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusCreated, product)
}
