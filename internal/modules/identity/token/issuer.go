package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the identity data embedded in an access token.
type Claims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"username"`
	jwt.RegisteredClaims
}

// Issuer signs access tokens for authenticated identities.
type Issuer struct {
	privateKey []byte
}

// NewIssuer creates an access-token issuer from an RSA private key in PEM format.
func NewIssuer(privateKey []byte) *Issuer {
	return &Issuer{privateKey: privateKey}
}

// Issue creates a 24-hour RS256 access token.
func (i *Issuer) Issue(userID, userName string) (string, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(i.privateKey)
	if err != nil {
		return "", fmt.Errorf("erro ao fazer parse de chave privada: %w", err)
	}

	claims := Claims{
		UserID:   userID,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("erro ao assinar token: %w", err)
	}

	return tokenString, nil
}
