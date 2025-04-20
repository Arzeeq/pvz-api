package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewJWTClaims(t *testing.T) {
	testcases := []struct {
		role     string
		duration time.Duration
	}{
		{
			role:     "moderator",
			duration: time.Hour,
		},
		{
			role:     "employee",
			duration: 24 * time.Hour,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.role, func(t *testing.T) {
			claims := NewJWTClaims(testcase.role, testcase.duration)
			duration := claims.ExpiresAt.Sub(claims.IssuedAt.Time)

			require.Equal(t, testcase.role, claims.Role)
			require.Equal(t, testcase.duration, duration)
		})
	}
}

func TestNewTokenService(t *testing.T) {
	testcases := []struct {
		name        string
		jwtSecret   []byte
		jwtDuration time.Duration
		expected    *TokenService
		err         error
	}{
		{
			name:        "success",
			jwtSecret:   []byte("test_secret"),
			jwtDuration: time.Hour,
			expected: &TokenService{
				jwtSecret:   []byte("test_secret"),
				jwtDuration: time.Hour,
			},
			err: nil,
		},
		{
			name:        "nil secret",
			jwtSecret:   nil,
			jwtDuration: time.Hour,
			expected:    nil,
			err:         ErrNilInConstruct,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			service, err := NewTokenService(testcase.jwtSecret, testcase.jwtDuration)
			require.ErrorIs(t, err, testcase.err)
			require.Equal(t, testcase.expected, service)
		})
	}
}

func TestGen(t *testing.T) {
	tokenService, err := NewTokenService([]byte("secret"), time.Hour)
	require.NoError(t, err)

	_, err = tokenService.Gen("role")
	require.NoError(t, err)
}
