package service

import (
	"context"
	"errors"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/pkg/auth"
)

var (
	ErrUserRegister    = errors.New("failed to register user")
	ErrUserLogin       = errors.New("failed to login user")
	ErrPasswordHashing = errors.New("failed to hash password")
	ErrTokenCreation   = errors.New("failed to cretate jwt token")
	ErrUserExists      = errors.New("user is already exists")
)

type UserStorager interface {
	CreateUser(ctx context.Context, payload dto.PostRegisterJSONBody) (*dto.User, error)
	GetUserPassword(ctx context.Context, email string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*dto.User, error)
}

type TokenServicer interface {
	Gen(role string) (dto.Token, error)
}

type UserService struct {
	storage      UserStorager
	tokenService TokenServicer
}

func NewUserService(storage UserStorager, tokenService TokenServicer) (*UserService, error) {
	if storage == nil || tokenService == nil {
		return nil, errors.New("nil values in NewUserService constructor")
	}

	return &UserService{storage: storage, tokenService: tokenService}, nil
}

func (s *UserService) RegisterUser(ctx context.Context, payload dto.PostRegisterJSONBody) (*dto.User, error) {
	hashedPassword, err := auth.HashPassword(payload.Password)
	if err != nil {
		return nil, ErrPasswordHashing
	}
	payload.Password = hashedPassword

	user, err := s.storage.CreateUser(ctx, payload)
	if err != nil {
		return nil, ErrUserExists
	}
	return user, nil
}

func (s *UserService) LoginUser(ctx context.Context, payload dto.PostLoginJSONBody) (dto.Token, error) {
	hashedPassword, err := s.storage.GetUserPassword(ctx, string(payload.Email))
	if err != nil {
		return "", ErrUserLogin
	}

	if !auth.ComparePasswords(hashedPassword, payload.Password) {
		return "", ErrUserLogin
	}

	user, err := s.storage.GetUserByEmail(ctx, string(payload.Email))
	if err != nil {
		return "", ErrUserLogin
	}

	token, err := s.tokenService.Gen(string(user.Role))
	if err != nil {
		return "", ErrTokenCreation
	}

	return token, nil
}
