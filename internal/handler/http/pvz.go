package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/go-playground/validator/v10"
)

type PVZServicer interface {
	CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error)
	GetPVZWithReceptions(ctx context.Context, payload dto.GetPvzParams) []dto.PVZWithReceptions
}

type PVZ struct {
	pvzService PVZServicer
	log        *logger.MyLogger
	validator  *validator.Validate
	timeout    time.Duration
}

func NewPvzHandler(pvzService PVZServicer, logger *logger.MyLogger, timeout time.Duration) (*PVZ, error) {
	if pvzService == nil || logger == nil {
		return nil, errors.New("nil values in NewPvzHandler constructor")
	}

	return &PVZ{
		pvzService: pvzService,
		log:        logger,
		validator:  validator.New(),
		timeout:    timeout,
	}, nil
}

func (h *PVZ) CreatePvz(w http.ResponseWriter, r *http.Request) {
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

func (h *PVZ) GetPVZ(w http.ResponseWriter, r *http.Request) {
	var pvzDto dto.GetPvzParams
	if err := pvzDto.FromParams(r); err != nil {
		h.log.HTTPResponse(w, http.StatusOK, nil)
		return
	}
	dto.CorrectParams(&pvzDto)

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	pvzs := h.pvzService.GetPVZWithReceptions(ctx, pvzDto)
	h.log.HTTPResponse(w, http.StatusOK, pvzs)
}
