package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims struct to hold the JWT claims
type Claims struct {
	ID         uint
	Name       string
	BufferTime int64
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a given username
func GenerateJWT(id uint, username string, expire int, key string) (string, error) {
	jwtKey := []byte(key)
	expirationTime := time.Now().Add(time.Second * time.Duration(expire))
	claims := &Claims{
		ID:         id,
		Name:       username,
		BufferTime: 3600,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "idb",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ValidateJWT validates the given JWT token and returns the claims if valid
func ValidateJWT(tokenString string, key string) (*Claims, error) {
	jwtKey := []byte(key)
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
