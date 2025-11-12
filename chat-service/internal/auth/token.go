package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

func ValidateToken(tokenString string) (*Claims, error) {
	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if publicKeyPath == "" {
		return nil, errors.New("WT_PUBLIC_KEY_PATH não está definido")
	}

	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("não foi possível ler a chave pública: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("não foi possível fazer o parse da chave pública: %w", err)
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", err)
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Agora só precisamos deste check
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// (Opcional) Adicionar verificação de expiração
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token expirado")
		}
		return claims, nil
	}
	return nil, errors.New("token inválido")
}
