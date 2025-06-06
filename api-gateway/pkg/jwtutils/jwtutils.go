package jwtutils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"user_name"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenString string) (*CustomClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET não configurado no ambiente")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token expirado")
		}
		return claims, nil
	}

	return nil, errors.New("token inválido")
}
