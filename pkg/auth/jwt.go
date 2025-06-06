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

func GetClaimsJWT(token string, secret []byte) (map[string]interface{}, error) {
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return secret, nil
	}

	parsedToken, err := jwt.Parse(token, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("token parsing error: %v", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	result := make(map[string]interface{})
	for k, v := range claims {
		result[k] = v
	}

	return result, nil
}
