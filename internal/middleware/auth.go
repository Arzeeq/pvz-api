package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/Arzeeq/pvz-api/internal/dto"
	"github.com/Arzeeq/pvz-api/internal/logger"
	"github.com/Arzeeq/pvz-api/pkg/auth"
)

var (
	ErrNoTokenProvided = errors.New("no token provided")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidRole     = errors.New("token has invalid role")
	ErrTokenExpired    = errors.New("token expired")
	ErrNoRoleProvided  = errors.New("no role provided")
	ErrNoExpProvided   = errors.New("no exp provided")
)

func AuthRoles(log *logger.MyLogger, jwtSecret []byte, roles ...dto.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.HTTPError(w, http.StatusForbidden, ErrNoTokenProvided)
				return
			}

			authParts := strings.Split(authHeader, " ")
			if len(authParts) != 2 || authParts[0] != "Bearer" {
				log.HTTPError(w, http.StatusForbidden, ErrInvalidToken)
				return
			}

			token := authParts[1]

			claims, err := auth.GetClaimsJWT(token, jwtSecret)
			if err != nil {
				log.HTTPError(w, http.StatusForbidden, ErrInvalidToken)
				return
			}

			if err := validateRole(claims, roles); err != nil {
				log.HTTPError(w, http.StatusForbidden, err)
				return
			}

			if err := validateExp(claims); err != nil {
				log.HTTPError(w, http.StatusForbidden, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateRole(claims map[string]interface{}, roles []dto.UserRole) error {
	authRole, ok := claims["role"].(string)
	if !ok {
		return ErrNoRoleProvided
	}

	for _, allowedRole := range roles {
		if dto.UserRole(authRole) == allowedRole {
			return nil
		}
	}

	return ErrInvalidRole
}

func validateExp(claims map[string]interface{}) error {
	exp, ok := claims["exp"].(float64)
	if !ok {
		return ErrNoExpProvided
	}

	expTime := time.Unix(int64(exp), 0)
	if time.Now().After(expTime) {
		return ErrTokenExpired
	}

	return nil
}
