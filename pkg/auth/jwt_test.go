package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestCreateJWTWithoutError(t *testing.T) {
	claims := struct {
		jwt.RegisteredClaims
		Payload string
	}{Payload: "some_payload"}

	_, err := CreateJWT([]byte("some_secret"), claims)

	require.NoError(t, err, "JWTCreate must execute without error")
}

func TestGetClaimsJWT(t *testing.T) {
	type Claims struct {
		jwt.RegisteredClaims
		Payload string
	}

	jwtDuration := time.Minute
	jwtSecret := []byte("some_secret")
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jwtDuration)),
		},
		Payload: "some_payload",
	}

	token, err := CreateJWT(jwtSecret, claims)
	require.NoError(t, err, "JWTCreate must execute without error")

	var newClaims Claims
	err = GetClaimsJWT(token, jwtSecret, &newClaims)
	require.NoError(t, err, "Failed to get claims")

	claimsDuration := claims.ExpiresAt.Sub(claims.IssuedAt.Time)

	require.Equal(t, jwtDuration, claimsDuration, "Incorrect duration")
	require.Equal(t, claims.Payload, newClaims.Payload, "Incorrect payload")
}
