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

type MockReceptionStorage struct {
	mock.Mock
}

func (m *MockReceptionStorage) CreateReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*dto.Reception), args.Error(1)
}

func (m *MockReceptionStorage) GetPVZReceptionsFiltered(ctx context.Context, pvzID openapi_types.UUID, startDate, endDate time.Time) []dto.Reception {
	args := m.Called(ctx, pvzID, startDate, endDate)
	return args.Get(0).([]dto.Reception)
}

func (m *MockReceptionStorage) GetActiveReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*dto.Reception), args.Error(1)
}

func (m *MockReceptionStorage) CloseReception(ctx context.Context, pvzID openapi_types.UUID) (*dto.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(*dto.Reception), args.Error(1)
}

func TestNewReceptionService(t *testing.T) {
	testcases := []struct {
		name    string
		storage ReceptionStorager
		wantErr bool
		err     error
	}{
		{
			name:    "success",
			storage: new(MockReceptionStorage),
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
			service, err := NewReceptionService(testcase.storage)
			require.ErrorIs(t, err, testcase.err)
			if testcase.err != nil {
				require.Nil(t, service)
			} else {
				require.NotNil(t, service)
			}
		})
	}
}

func TestReceptionService_CreateReception(t *testing.T) {
	ctx := context.Background()
	testUUID := openapi_types.UUID{}

	testcases := []struct {
		name      string
		pvzID     openapi_types.UUID
		mockSetup func(*MockReceptionStorage)
		expected  *dto.Reception
		err       error
	}{
		{
			name:  "successful creation",
			pvzID: testUUID,
			mockSetup: func(m *MockReceptionStorage) {
				m.On("CreateReception", ctx, testUUID).
					Return(&dto.Reception{
						Id:     testUUID,
						PvzId:  testUUID,
						Status: dto.InProgress,
					}, nil)
			},
			expected: &dto.Reception{
				Id:     testUUID,
				PvzId:  testUUID,
				Status: dto.InProgress,
			},
			err: nil,
		},
		{
			name:  "storage error",
			pvzID: testUUID,
			mockSetup: func(m *MockReceptionStorage) {
				m.On("CreateReception", ctx, testUUID).
					Return(&dto.Reception{}, errors.New("storage error"))
			},
			expected: nil,
			err:      ErrReceptionCreate,
		},
		{
			name:  "active reception exists",
			pvzID: testUUID,
			mockSetup: func(m *MockReceptionStorage) {
				m.On("CreateReception", ctx, testUUID).
					Return(&dto.Reception{}, ErrActiveReception)
			},
			expected: nil,
			err:      ErrReceptionCreate,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			mockStorage := new(MockReceptionStorage)
			testcase.mockSetup(mockStorage)

			service, err := NewReceptionService(mockStorage)
			require.NoError(t, err)
			reception, err := service.CreateReception(ctx, testcase.pvzID)

			require.Equal(t, testcase.expected, reception)
			require.Equal(t, testcase.err, err)
			mockStorage.AssertExpectations(t)
		})
	}
}

func TestReceptionService_CloseReception(t *testing.T) {
	ctx := context.Background()
	testUUID := openapi_types.UUID{}

	tests := []struct {
		name      string
		pvzID     openapi_types.UUID
		mockSetup func(*MockReceptionStorage)
		expected  *dto.Reception
		err       error
	}{
		{
			name:  "successful close",
			pvzID: testUUID,
			mockSetup: func(m *MockReceptionStorage) {
				m.On("CloseReception", ctx, testUUID).
					Return(&dto.Reception{
						Id:     testUUID,
						PvzId:  testUUID,
						Status: dto.Close,
					}, nil)
			},
			expected: &dto.Reception{
				Id:     testUUID,
				PvzId:  testUUID,
				Status: dto.Close,
			},
			err: nil,
		},
		{
			name:  "storage error",
			pvzID: testUUID,
			mockSetup: func(m *MockReceptionStorage) {
				m.On("CloseReception", ctx, testUUID).
					Return(&dto.Reception{}, errors.New("storage error"))
			},
			expected: nil,
			err:      ErrReceptionClose,
		},
		{
			name:  "no active reception to close",
			pvzID: testUUID,
			mockSetup: func(m *MockReceptionStorage) {
				m.On("CloseReception", ctx, testUUID).
					Return(&dto.Reception{}, errors.New("no active reception"))
			},
			expected: nil,
			err:      ErrReceptionClose,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockReceptionStorage)
			tt.mockSetup(mockStorage)

			service, err := NewReceptionService(mockStorage)
			require.NoError(t, err)
			reception, err := service.CloseReception(ctx, tt.pvzID)

			require.Equal(t, tt.expected, reception)
			require.Equal(t, tt.err, err)
			mockStorage.AssertExpectations(t)
		})
	}
}
