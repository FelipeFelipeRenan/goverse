package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	jwt.RegisteredClaims
}

func main() {
	http.HandleFunc("/auth/validate", validateHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("Auth middleware rodando na porta: ", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func validateHandler(w http.ResponseWriter, r *http.Request) {
	// Adiciona CORS SEMPRE
	addCORSHeaders(w)
	// Tratar pré-flight OPTIONS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	token := extractTokenFromHeader(r)
	if token == "" {
		respondUnauthorized(w, "Token não fornecido")
		return
	}

	claims, err := validateToken(token)
	if err != nil {
		respondUnauthorized(w, "Token inválido: "+err.Error())
		return
	}
	if claims.UserID == "" {
		respondUnauthorized(w, "Token sem user_id")
		return
	}

	w.Header().Set("X-User-ID", claims.UserID)
	w.WriteHeader(http.StatusOK)
}

func addCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Ajuste se quiser restringir
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Vary", "Origin")
}

func respondUnauthorized(w http.ResponseWriter, msg string) {
	addCORSHeaders(w)
	http.Error(w, msg, http.StatusUnauthorized)
}

func extractTokenFromHeader(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}

	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

func validateToken(tokenString string) (*CustomClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET não configurado no ambiente")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token expirado")
		}
		return claims, nil
	}
	return nil, errors.New("token inválido")
}
