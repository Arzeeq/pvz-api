package service

import (
	"context"
	"errors"

	"github.com/Arzeeq/pvz-api/internal/dto"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var ErrProductCreate = errors.New("failed to create product")
var ErrDeleteProduct = errors.New("failed to delete product")

type ProductStorager interface {
	CreateProduct(ctx context.Context, productDto dto.PostProductsJSONBody) (*dto.Product, error)
	DeleteProduct(ctx context.Context, productID openapi_types.UUID) error
	GetLastProduct(ctx context.Context, pvzId openapi_types.UUID) (*dto.Product, error)
	GetReceptionProducts(ctx context.Context, receptionId openapi_types.UUID) []dto.Product
}

type ProductService struct {
	storage ProductStorager
}

func NewProductService(productStorage ProductStorager) (*ProductService, error) {
	if productStorage == nil {
		return nil, ErrNilInConstruct
	}

	return &ProductService{storage: productStorage}, nil
}

func (s *ProductService) CreateProduct(ctx context.Context, productDto dto.PostProductsJSONBody) (*dto.Product, error) {
	product, err := s.storage.CreateProduct(ctx, productDto)
	if err != nil {
		return nil, ErrProductCreate
	}

	return product, nil
}

func (s *ProductService) DeleteLastProduct(ctx context.Context, pvzID openapi_types.UUID) error {
	product, err := s.storage.GetLastProduct(ctx, pvzID)
	if err != nil {
		return ErrDeleteProduct
	}

	if err := s.storage.DeleteProduct(ctx, *product.Id); err != nil {
		return ErrDeleteProduct
	}

	return nil
}
