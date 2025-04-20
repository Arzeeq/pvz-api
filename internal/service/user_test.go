package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/pkg/auth"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockUserStorage struct {
	mock.Mock
}

func (m *MockUserStorage) CreateUser(ctx context.Context, payload dto.PostRegisterJSONBody) (*dto.User, error) {
	args := m.Called(ctx, payload)
	return args.Get(0).(*dto.User), args.Error(1)
}

func (m *MockUserStorage) GetUserPassword(ctx context.Context, email string) (string, error) {
	args := m.Called(ctx, email)
	return args.String(0), args.Error(1)
}

func (m *MockUserStorage) GetUserByEmail(ctx context.Context, email string) (*dto.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*dto.User), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) Gen(role string) (dto.Token, error) {
	args := m.Called(role)
	return args.Get(0).(dto.Token), args.Error(1)
}

func TestNewUserService(t *testing.T) {
	testcases := []struct {
		name         string
		tokenService TokenServicer
		userStorage  UserStorager
		err          error
		isNil        bool
	}{
		{
			name:         "success",
			tokenService: new(MockTokenService),
			userStorage:  new(MockUserStorage),
			err:          nil,
			isNil:        false,
		},
		{
			name:         "nil token service",
			tokenService: nil,
			userStorage:  new(MockUserStorage),
			err:          ErrNilInConstruct,
			isNil:        true,
		},
		{
			name:         "nil user storage",
			tokenService: new(MockTokenService),
			userStorage:  nil,
			err:          ErrNilInConstruct,
			isNil:        true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			service, err := NewUserService(testcase.userStorage, testcase.tokenService)
			require.ErrorIs(t, err, testcase.err)
			if testcase.isNil {
				require.Nil(t, service)
			} else {
				require.NotNil(t, service)
			}
		})
	}
}

func TestUserService_RegisterUser(t *testing.T) {
	ctx := context.Background()
	payload := dto.PostRegisterJSONBody{
		Email:    "test@example.com",
		Password: "password",
		Role:     "user",
	}
	expectedUser := &dto.User{
		Id:    &openapi_types.UUID{},
		Email: "test@example.com",
		Role:  "user",
	}

	storage1 := new(MockUserStorage)
	storage2 := new(MockUserStorage)
	storage1.On("CreateUser", ctx, mock.Anything).Return(expectedUser, nil)
	storage2.On("CreateUser", ctx, mock.Anything).Return(&dto.User{}, ErrUserExists)

	for _, testcase := range []struct {
		name         string
		storage      *MockUserStorage
		tokenService TokenServicer
		expectedUser *dto.User
		err          error
	}{
		{
			name:         "success",
			storage:      storage1,
			tokenService: new(MockTokenService),
			expectedUser: expectedUser,
			err:          nil,
		},
		{
			name:         "user exists",
			storage:      storage2,
			tokenService: new(MockTokenService),
			expectedUser: nil,
			err:          ErrUserExists,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			service := &UserService{
				storage:      testcase.storage,
				tokenService: testcase.tokenService,
			}
			user, err := service.RegisterUser(ctx, payload)
			require.ErrorIs(t, err, testcase.err)
			require.Equal(t, user, testcase.expectedUser)

			testcase.storage.AssertExpectations(t)
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	ctx := context.Background()
	password := "password"
	payload := dto.PostLoginJSONBody{
		Email:    "test@example.com",
		Password: password,
	}

	expectedUser := &dto.User{
		Id:    &openapi_types.UUID{},
		Email: "test@example.com",
		Role:  "user",
	}

	generatedToken := "token"

	hashedPassword, err := auth.HashPassword(password)
	require.NoError(t, err)

	tokenService := new(MockTokenService)
	tokenService.On("Gen", string(expectedUser.Role)).Return(generatedToken, nil)

	tokenServiceWithError := new(MockTokenService)
	tokenServiceWithError.On("Gen", string(expectedUser.Role)).Return("", errors.New("error"))

	for _, testcase := range []struct {
		name         string
		storageSetup func(*MockUserStorage)
		tokenService *MockTokenService
		token        string
		err          error
	}{
		{
			name: "success",
			storageSetup: func(m *MockUserStorage) {
				m.On("GetUserPassword", ctx, string(payload.Email)).Return(hashedPassword, nil)
				m.On("GetUserByEmail", ctx, string(payload.Email)).Return(expectedUser, nil)
			},
			tokenService: tokenService,
			token:        generatedToken,
			err:          nil,
		},
		{
			name: "get password error",
			storageSetup: func(m *MockUserStorage) {
				m.On("GetUserPassword", ctx, string(payload.Email)).Return("", errors.New("error"))
			},
			tokenService: tokenService,
			token:        "",
			err:          ErrUserLogin,
		},
		{
			name: "password mismatch",
			storageSetup: func(m *MockUserStorage) {
				m.On("GetUserPassword", ctx, string(payload.Email)).Return("wrong_hash", nil)
			},
			tokenService: tokenService,
			token:        "",
			err:          ErrUserLogin,
		},
		{
			name: "get user error",
			storageSetup: func(m *MockUserStorage) {
				m.On("GetUserPassword", ctx, string(payload.Email)).Return(hashedPassword, nil)
				m.On("GetUserByEmail", ctx, string(payload.Email)).Return(&dto.User{}, errors.New("error"))
			},
			tokenService: tokenService,
			token:        "",
			err:          ErrUserLogin,
		},
		{
			name: "token generation error",
			storageSetup: func(m *MockUserStorage) {
				m.On("GetUserPassword", ctx, string(payload.Email)).Return(hashedPassword, nil)
				m.On("GetUserByEmail", ctx, string(payload.Email)).Return(expectedUser, nil)
			},
			tokenService: tokenServiceWithError,
			token:        "",
			err:          ErrTokenCreation,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			storage := new(MockUserStorage)
			testcase.storageSetup(storage)
			service := &UserService{
				storage:      storage,
				tokenService: testcase.tokenService,
			}

			token, err := service.LoginUser(ctx, payload)
			require.ErrorIs(t, err, testcase.err)
			require.Equal(t, testcase.token, token)

			storage.AssertExpectations(t)
			testcase.tokenService.AssertExpectations(t)
		})
	}
}
