package service

import (
	"context"
	"errors"

	"github.com/Arzeeq/pvz-api/internal/dto"
)

var ErrPVZCreate = errors.New("failed to create PVZ")

type PVZStorager interface {
	CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error)
	GetPVZs(ctx context.Context, params dto.GetPvzParams) ([]dto.PVZ, error)
}

type PVZService struct {
	pvzStorage       PVZStorager
	receptionStorage ReceptionStorager
	productStorage   ProductStorager
}

func NewPVZService(
	pvzStorage PVZStorager,
	receptionStorage ReceptionStorager,
	productStorage ProductStorager,
) (*PVZService, error) {
	if pvzStorage == nil || receptionStorage == nil || productStorage == nil {
		return nil, ErrNilInConstruct
	}

	return &PVZService{
		pvzStorage:       pvzStorage,
		receptionStorage: receptionStorage,
		productStorage:   productStorage,
	}, nil
}

func (s *PVZService) CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error) {
	pvz, err := s.pvzStorage.CreatePVZ(ctx, payload)
	if err != nil {
		return nil, ErrPVZCreate
	}

	return pvz, nil
}

func (s *PVZService) GetPVZWithReceptionsFiltered(ctx context.Context, payload dto.GetPvzParams) []dto.PVZWithReceptions {
	pvzs, err := s.pvzStorage.GetPVZs(ctx, payload)
	if err != nil {
		return nil
	}

	result := make([]dto.PVZWithReceptions, 0)
	for i := range pvzs {
		receptions := s.receptionStorage.GetPVZReceptionsFiltered(
			ctx,
			*pvzs[i].Id,
			*payload.StartDate,
			*payload.EndDate,
		)

		pvzWithReceptions := dto.PVZWithReceptions{Pvz: &pvzs[i]}
		for j := range receptions {
			products := s.productStorage.GetReceptionProducts(ctx, receptions[j].Id)
			receptionWithProducts := dto.ReceptionWithProducts{
				Reception: receptions[j],
				Products:  products,
			}
			pvzWithReceptions.Receptions = append(pvzWithReceptions.Receptions, receptionWithProducts)
		}
		result = append(result, pvzWithReceptions)
	}

	return result
}
