package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

func CreateJWT(secret []byte, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func GetClaimsJWT(token string, secret []byte, claims jwt.Claims) error {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return secret, nil
	}

	parsedToken, err := jwt.ParseWithClaims(token, claims, keyFunc)
	if err != nil {
		return fmt.Errorf("token parsing error: %v", err)
	}

	if !parsedToken.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}
