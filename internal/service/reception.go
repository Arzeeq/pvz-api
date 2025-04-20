package service

import (
	"context"
	"errors"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrActiveReception = errors.New("failed to create reception, there is already an active reception")
var ErrReceptionCreate = errors.New("failed to create reception")
var ErrReceptionClose = errors.New("failed to close reception")

type ReceptionStorager interface {
	CreateReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error)
	GetPVZReceptionsFiltered(ctx context.Context, pvzID openapi_types.UUID, startDate, endDate time.Time) []dto.Reception
	CloseReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error)
}

type ReceptionService struct {
	storage ReceptionStorager
}

func NewReceptionService(storage ReceptionStorager) (*ReceptionService, error) {
	if storage == nil {
		return nil, ErrNilInConstruct
	}

	return &ReceptionService{storage: storage}, nil
}

func (s *ReceptionService) CreateReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	reception, err := s.storage.CreateReception(ctx, pvzID)
	if err != nil {
		return nil, ErrReceptionCreate
	}

	return reception, nil
}

func (s *ReceptionService) CloseReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	reception, err := s.storage.CloseReception(ctx, pvzID)
	if err != nil {
		return nil, ErrReceptionClose
	}

	return reception, nil
}
