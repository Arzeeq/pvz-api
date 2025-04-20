package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockProductStorage struct {
	mock.Mock
}

func (m *MockProductStorage) CreateProduct(ctx context.Context, productDto dto.PostProductsJSONBody) (*dto.Product, error) {
	args := m.Called(ctx, productDto)
	return args.Get(0).(*dto.Product), args.Error(1)
}

func (m *MockProductStorage) DeleteProduct(ctx context.Context, productID openapi_types.UUID) error {
	args := m.Called(ctx, productID)
	return args.Error(0)
}

func (m *MockProductStorage) GetLastProduct(ctx context.Context, pvzId openapi_types.UUID) (*dto.Product, error) {
	args := m.Called(ctx, pvzId)
	return args.Get(0).(*dto.Product), args.Error(1)
}

func (m *MockProductStorage) GetReceptionProducts(ctx context.Context, receptionId openapi_types.UUID) []dto.Product {
	args := m.Called(ctx, receptionId)
	return args.Get(0).([]dto.Product)
}

func TestNewProductService(t *testing.T) {
	testcases := []struct {
		name    string
		storage ProductStorager
		err     error
	}{
		{
			name:    "successful initialization",
			storage: new(MockProductStorage),
			err:     nil,
		},
		{
			name:    "nil storage",
			storage: nil,
			err:     ErrNilInConstruct,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			service, err := NewProductService(testcase.storage)
			require.ErrorIs(t, err, testcase.err)
			if testcase.err != nil {
				require.Nil(t, service)
			} else {
				require.NotNil(t, service)
			}
		})
	}
}

func TestProductService_CreateProduct(t *testing.T) {
	ctx := context.Background()
	testUUID := openapi_types.UUID{}
	receptionUUID := openapi_types.UUID{}
	pvzID := openapi_types.UUID{}
	now := time.Now()
	productType := dto.ProductTypeClothes

	productDto := dto.PostProductsJSONBody{
		PvzId: pvzID,
		Type:  dto.PostProductsJSONBodyTypeClothes,
	}

	testcases := []struct {
		name      string
		input     dto.PostProductsJSONBody
		mockSetup func(*MockProductStorage)
		expected  *dto.Product
		err       error
	}{
		{
			name:  "successful product creation",
			input: productDto,
			mockSetup: func(m *MockProductStorage) {
				m.On("CreateProduct", ctx, productDto).
					Return(&dto.Product{
						Id:          &testUUID,
						DateTime:    &now,
						ReceptionId: receptionUUID,
						Type:        productType,
					}, nil)
			},
			expected: &dto.Product{
				Id:          &testUUID,
				DateTime:    &now,
				ReceptionId: receptionUUID,
				Type:        productType,
			},
			err: nil,
		},
		{
			name:  "storage error on create",
			input: productDto,
			mockSetup: func(m *MockProductStorage) {
				m.On("CreateProduct", ctx, productDto).
					Return(&dto.Product{}, errors.New("storage error"))
			},
			expected: nil,
			err:      ErrProductCreate,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			storage := new(MockProductStorage)
			testcase.mockSetup(storage)

			service, err := NewProductService(storage)
			require.NoError(t, err)
			fmt.Println(service, testcase)
			product, err := service.CreateProduct(ctx, testcase.input)

			require.Equal(t, testcase.expected, product)
			require.ErrorIs(t, testcase.err, err)
			storage.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteLastProduct(t *testing.T) {
	ctx := context.Background()
	pvzID := openapi_types.UUID{}
	productID := openapi_types.UUID{}
	receptionID := openapi_types.UUID{}
	now := time.Now()
	productType := dto.ProductTypeClothes

	testcases := []struct {
		name      string
		pvzID     openapi_types.UUID
		mockSetup func(*MockProductStorage)
		err       error
	}{
		{
			name:  "successful deletion",
			pvzID: pvzID,
			mockSetup: func(m *MockProductStorage) {
				m.On("GetLastProduct", ctx, pvzID).
					Return(&dto.Product{
						Id:          &productID,
						DateTime:    &now,
						ReceptionId: receptionID,
						Type:        productType,
					}, nil)
				m.On("DeleteProduct", ctx, productID).
					Return(nil)
			},
			err: nil,
		},
		{
			name:  "error getting last product",
			pvzID: pvzID,
			mockSetup: func(m *MockProductStorage) {
				m.On("GetLastProduct", ctx, pvzID).
					Return(&dto.Product{}, errors.New("error"))
			},
			err: ErrDeleteProduct,
		},
		{
			name:  "error deleting product",
			pvzID: pvzID,
			mockSetup: func(m *MockProductStorage) {
				m.On("GetLastProduct", ctx, pvzID).
					Return(&dto.Product{
						Id:          &productID,
						DateTime:    &now,
						ReceptionId: receptionID,
						Type:        productType,
					}, nil)
				m.On("DeleteProduct", ctx, productID).
					Return(errors.New("delete failed"))
			},
			err: ErrDeleteProduct,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			storage := new(MockProductStorage)
			testcase.mockSetup(storage)

			service, err := NewProductService(storage)
			require.NoError(t, err)
			err = service.DeleteLastProduct(ctx, testcase.pvzID)

			require.ErrorIs(t, err, testcase.err)
			storage.AssertExpectations(t)
		})
	}
}
