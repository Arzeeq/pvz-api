package service

import (
	"context"
	"errors"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrPVZCreate = errors.New("failed to create PVZ")

type PVZStorager interface {
	CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error)
	GetPVZ(ctx context.Context, page, limit int) ([]dto.PVZ, error)
}

type ReceptionStorager interface {
	GetPVZReceptionsFiltered(ctx context.Context, pvzID openapi_types.UUID, startDate, endDate time.Time) ([]dto.Reception, error)
}

type ProductStorager interface {
	GetReceptionProducts(ctx context.Context, receptionId openapi_types.UUID) ([]dto.Product, error)
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
		return nil, errors.New("nil values in NewPvzService constructor")
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

func (s *PVZService) GetPVZWithReceptions(ctx context.Context, payload dto.GetPvzParams) []dto.PVZWithReceptions {
	pvzs, err := s.pvzStorage.GetPVZ(ctx, *payload.Page, *payload.Limit)

	if err != nil {
		return nil
	}

	var result []dto.PVZWithReceptions
	for i := range pvzs {
		receptions, err := s.receptionStorage.GetPVZReceptionsFiltered(ctx, *pvzs[i].Id, *payload.StartDate, *payload.EndDate)
		pvzWithRecepctions := dto.PVZWithReceptions{PVZ: pvzs[i]}
		if err == nil {
			var receptionsWithProducts []dto.ReceptionWithProducts
			for j := range receptions {
				products, err := s.productStorage.GetReceptionProducts(ctx, *receptions[j].Id)
				receptionWithProducts := dto.ReceptionWithProducts{Reception: receptions[j]}
				if err == nil {
					receptionWithProducts.Products = products
				}

				receptionsWithProducts = append(receptionsWithProducts, receptionWithProducts)
			}

			pvzWithRecepctions.Receptions = receptionsWithProducts
		}

		result = append(result, pvzWithRecepctions)
	}

	return result
}
