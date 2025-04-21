package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockPVZStorage struct {
	mock.Mock
}

func (m *mockPVZStorage) CreatePVZ(ctx context.Context, payload dto.PostPvzJSONRequestBody) (*dto.PVZ, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(*dto.PVZ), args.Error(1)
}

func (m *mockPVZStorage) GetPVZs(ctx context.Context, params dto.GetPvzParams) ([]dto.PVZ, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]dto.PVZ), args.Error(1)
}

func (m *mockPVZStorage) GetAllPVZs(ctx context.Context) []dto.PVZ {
	return nil
}

func TestPVZService_CreatePVZ(t *testing.T) {
	// preparaion
	now := time.Now()
	expected := &dto.PVZ{
		Id:               &openapi_types.UUID{},
		City:             dto.Moscow,
		RegistrationDate: &now,
	}
	input := dto.PostPvzJSONRequestBody{City: dto.Moscow}

	// testcases
	testcases := []struct {
		name      string
		mockSetup func(*mockPVZStorage)
		input     dto.PostPvzJSONRequestBody
		expected  *dto.PVZ
		err       error
	}{
		{
			name: "successful creation",
			mockSetup: func(m *mockPVZStorage) {
				m.On("CreatePVZ", mock.Anything, input).
					Return(expected, nil)
			},
			input:    input,
			expected: expected,
			err:      nil,
		},
		{
			name: "storage error",
			mockSetup: func(m *mockPVZStorage) {
				m.On("CreatePVZ", mock.Anything, mock.Anything).
					Return(&dto.PVZ{}, errors.New("error"))
			},
			input:    input,
			expected: nil,
			err:      ErrPVZCreate,
		},
	}

	// run tests
	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			pvzStorageMock := &mockPVZStorage{}
			testcase.mockSetup(pvzStorageMock)

			service, err := NewPVZService(
				pvzStorageMock,
				&mockReceptionStorage{},
				&mockProductStorage{},
			)
			require.NoError(t, err)

			result, err := service.CreatePVZ(context.Background(), testcase.input)

			if testcase.err != nil {
				require.ErrorIs(t, err, testcase.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.expected.Id, result.Id)
				require.Equal(t, testcase.expected.City, result.City)
				require.Equal(t, testcase.expected.RegistrationDate, result.RegistrationDate)
			}
			pvzStorageMock.AssertExpectations(t)
		})
	}
}

func TestPVZService_GetPVZWithReceptionsFiltered(t *testing.T) {
	// preparation
	now := time.Now()
	pvzID := openapi_types.UUID{}
	receptionID := openapi_types.UUID{}
	page := 1
	limit := 10
	input := dto.GetPvzParams{
		StartDate: &now,
		EndDate:   &now,
		Page:      &page,
		Limit:     &limit,
	}

	// testcases
	testcases := []struct {
		name        string
		mockSetup   func(*mockPVZStorage, *mockReceptionStorage, *mockProductStorage)
		input       dto.GetPvzParams
		expected    []dto.PVZWithReceptions
		expectError bool
	}{
		{
			name: "successful get with receptions and products",
			mockSetup: func(p *mockPVZStorage, r *mockReceptionStorage, pr *mockProductStorage) {
				// Mock PVZs
				p.On("GetPVZs", mock.Anything, dto.GetPvzParams{
					StartDate: &now,
					EndDate:   &now,
					Page:      &page,
					Limit:     &limit,
				}).Return([]dto.PVZ{
					{
						Id:               &pvzID,
						City:             dto.Moscow,
						RegistrationDate: &now,
					},
				}, nil)

				// Mock Receptions
				r.On("GetPVZReceptionsFiltered", mock.Anything, pvzID, now, now).
					Return([]dto.Reception{
						{
							Id:       receptionID,
							PvzId:    pvzID,
							DateTime: now,
							Status:   dto.InProgress,
						},
					}, nil)

				// Mock Products
				pr.On("GetReceptionProducts", mock.Anything, receptionID).
					Return([]dto.Product{
						{
							Id:          &openapi_types.UUID{},
							Type:        dto.ProductTypeElectronics,
							ReceptionId: receptionID,
							DateTime:    &now,
						},
					}, nil)
			},
			input: input,
			expected: []dto.PVZWithReceptions{
				{
					Pvz: &dto.PVZ{
						Id:               &pvzID,
						City:             dto.Moscow,
						RegistrationDate: &now,
					},
					Receptions: []dto.ReceptionWithProducts{
						{
							Reception: dto.Reception{
								Id:       receptionID,
								PvzId:    pvzID,
								DateTime: now,
								Status:   dto.InProgress,
							},
							Products: []dto.Product{
								{
									Id:          &openapi_types.UUID{},
									Type:        dto.ProductTypeElectronics,
									ReceptionId: receptionID,
									DateTime:    &now,
								},
							},
						},
					},
				},
			},
			expectError: false,
		},
		{
			name: "error getting PVZs",
			mockSetup: func(p *mockPVZStorage, r *mockReceptionStorage, pr *mockProductStorage) {
				p.On("GetPVZs", mock.Anything, mock.Anything).Return([]dto.PVZ{}, errors.New("error"))
			},
			input:       dto.GetPvzParams{},
			expected:    nil,
			expectError: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			// arrange
			pvzStorageMock := &mockPVZStorage{}
			receptionStorageMock := &mockReceptionStorage{}
			productStorageMock := &mockProductStorage{}
			testcase.mockSetup(pvzStorageMock, receptionStorageMock, productStorageMock)

			service, err := NewPVZService(
				pvzStorageMock,
				receptionStorageMock,
				productStorageMock,
			)
			require.NoError(t, err)

			// act
			result := service.GetPVZWithReceptionsFiltered(context.Background(), testcase.input)

			// assert
			if testcase.expectError {
				require.Nil(t, result)
			} else {
				require.Equal(t, len(testcase.expected), len(result))
				if len(result) > 0 {
					require.Equal(t, testcase.expected[0].Pvz.Id, result[0].Pvz.Id)
					require.Equal(t, len(testcase.expected[0].Receptions), len(result[0].Receptions))
					if len(result[0].Receptions) > 0 {
						require.Equal(t, testcase.expected[0].Receptions[0].Reception.Id, result[0].Receptions[0].Reception.Id)
						require.Equal(t, len(testcase.expected[0].Receptions[0].Products), len(result[0].Receptions[0].Products))
					}
				}
			}

			pvzStorageMock.AssertExpectations(t)
			receptionStorageMock.AssertExpectations(t)
			productStorageMock.AssertExpectations(t)
		})
	}
}
