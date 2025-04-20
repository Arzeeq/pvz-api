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

type ReceptionServicer interface {
	CreateReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error)
	CloseReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error)
}

type ReceptionHandler struct {
	receptionService ReceptionServicer
	log              *logger.MyLogger
	validator        *validator.Validate
	timeout          time.Duration
}

func NewReceptionHandler(receptionService ReceptionServicer, logger *logger.MyLogger, timeout time.Duration) (*ReceptionHandler, error) {
	if receptionService == nil || logger == nil {
		return nil, errors.New("nil values in NewPvzHandler constructor")
	}

	return &ReceptionHandler{
		receptionService: receptionService,
		log:              logger,
		validator:        validator.New(),
		timeout:          timeout,
	}, nil
}

func (h *ReceptionHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	var receptionDto dto.PostReceptionsJSONBody
	if err := dto.Parse(r.Body, &receptionDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	user, err := h.receptionService.CreateReception(ctx, receptionDto.PvzId)
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusCreated, user)
}
