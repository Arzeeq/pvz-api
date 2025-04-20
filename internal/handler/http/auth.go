package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/go-playground/validator/v10"
)

var ErrValidationFailed = errors.New("failed validation of request body")

type UserServicer interface {
	RegisterUser(ctx context.Context, payload dto.PostRegisterJSONBody) (*dto.User, error)
	LoginUser(ctx context.Context, payload dto.PostLoginJSONBody) (dto.Token, error)
}

type TokenServicer interface {
	Gen(role string) (dto.Token, error)
}

type AuthHandler struct {
	userService  UserServicer
	tokenService TokenServicer
	log          *logger.MyLogger
	timeout      time.Duration
	validator    *validator.Validate
}

func NewAuthHandler(
	userService UserServicer,
	tokenService TokenServicer,
	logger *logger.MyLogger,
	timeout time.Duration,
) (*AuthHandler, error) {
	if userService == nil || tokenService == nil || logger == nil {
		return nil, errors.New("nil pointers in NewAuthHandler constructor")
	}

	return &AuthHandler{
		userService:  userService,
		tokenService: tokenService,
		log:          logger,
		timeout:      timeout,
		validator:    validator.New(),
	}, nil
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
	var roleDto dto.PostDummyLoginJSONBody
	if err := dto.Parse(r.Body, &roleDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validator.Struct(roleDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, ErrValidationFailed)
		return
	}

	token, err := h.tokenService.Gen(string(roleDto.Role))
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusOK, token)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var userDto dto.PostRegisterJSONBody
	if err := dto.Parse(r.Body, &userDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}
	if err := h.validator.Struct(userDto); err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, ErrValidationFailed)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	user, err := h.userService.RegisterUser(ctx, userDto)
	if err != nil {
		h.log.HTTPError(w, http.StatusBadRequest, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusCreated, user)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var userDto dto.PostLoginJSONBody
	if err := dto.Parse(r.Body, &userDto); err != nil {
		h.log.HTTPError(w, http.StatusUnauthorized, err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	token, err := h.userService.LoginUser(ctx, userDto)
	if err != nil {
		h.log.HTTPError(w, http.StatusUnauthorized, err)
		return
	}

	h.log.HTTPResponse(w, http.StatusOK, token)
}
