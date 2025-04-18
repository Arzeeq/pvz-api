package service

import (
	"errors"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
)

var ErrTokenGen = errors.New("failed to generate token")

type JWTClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func NewJWTClaims(role string, duration time.Duration) *JWTClaims {
	now := time.Now()
	return &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
		Role: role,
	}
}

type TokenService struct {
	jwtSecret   []byte
	jwtDuration time.Duration
}

func NewTokenService(jwtSecret string, jwtDuration time.Duration) *TokenService {
	return &TokenService{jwtSecret: []byte(jwtSecret), jwtDuration: jwtDuration}
}

func (s *TokenService) Gen(role string) (dto.Token, error) {
	claims := NewJWTClaims(role, s.jwtDuration)
	token, err := auth.CreateJWT(s.jwtSecret, claims)
	if err != nil {
		return "", ErrTokenCreation
	}

	return token, nil
}
