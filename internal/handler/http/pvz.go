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

type PVZServicer interface {
	CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error)
	GetPVZWithReceptionsFiltered(ctx context.Context, payload dto.GetPvzParams) []dto.PVZWithReceptions
}

type PVZHandler struct {
	pvzService       PVZServicer
	receptionService ReceptionServicer
	productService   ProductServicer
	log              *logger.MyLogger
	validator        *validator.Validate
	timeout          time.Duration
}

func NewPvzHandler(
	pvzService PVZServicer,
	receptionService ReceptionServicer,
	productService ProductServicer,
	logger *logger.MyLogger,
	timeout time.Duration,
) (*PVZHandler, error) {
	if pvzService == nil || receptionService == nil || productService == nil || logger == nil {
		return nil, errors.New("nil values in NewPvzHandler constructor")
	}

	return &PVZHandler{
		pvzService:       pvzService,
		receptionService: receptionService,
		productService:   productService,
		log:              logger,
		validator:        validator.New(),
		timeout:          timeout,
	}, nil
}

func (h *PVZHandler) CreatePvz(w http.ResponseWriter, r *http.Request) {
	var pvzDto dto.PostPvzJSONRequestBody
	if err := dto.Parse(r.Body, &pvzDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validator.Struct(pvzDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, ErrValidationFailed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	pvz, err := h.pvzService.CreatePVZ(ctx, pvzDto)
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusCreated, pvz)
}

func (h *PVZHandler) GetPVZ(w http.ResponseWriter, r *http.Request) {
	var pvzDto dto.GetPvzParams
	if err := pvzDto.FromParams(r); err != nil {
		h.log.HTTPResponse(w, http.StatusOK, nil)
		return
	}
	dto.CorrectParams(&pvzDto)

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	pvzs := h.pvzService.GetPVZWithReceptionsFiltered(ctx, pvzDto)
	h.log.HTTPResponse(w, http.StatusOK, pvzs)
}

func (h *PVZHandler) CloseReception(w http.ResponseWriter, r *http.Request) {
	pathValue := r.PathValue("pvzId")
	var pvzId openapi_types.UUID
	err := pvzId.UnmarshalText([]byte(pathValue))
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	reception, err := h.receptionService.CloseReception(ctx, pvzId)
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusOK, reception)
}

func (h *PVZHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	pathValue := r.PathValue("pvzId")
	var pvzId openapi_types.UUID
	err := pvzId.UnmarshalText([]byte(pathValue))
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	err = h.productService.DeleteLastProduct(ctx, pvzId)
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}
}
