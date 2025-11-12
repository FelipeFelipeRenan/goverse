package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
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

	http.HandleFunc("/health", healthCheck)

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

	cookie, err := r.Cookie("access_token")
	if err != nil {
		if err == http.ErrNoCookie {
			respondUnauthorized(w, "Cookie de atuenticação não encontrado")
			return
		}
		respondUnauthorized(w, "Requisição inválida")
		return
	}

	tokenString := cookie.Value

	claims, err := validateToken(tokenString)
	if err != nil {
		log.Printf("ERRO DE VALIDAÇÃO DO TOKEN: %v", err)

		respondUnauthorized(w, "Token inválido: "+err.Error())
		return
	}
	if claims.UserID == "" {
		respondUnauthorized(w, "Token sem user_id")
		return
	}

	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE" {
		csrfCookie, err := r.Cookie("csrf_token")
		if err != nil {
			respondUnauthorized(w, "Cookie CSRF não encontrado")
			return
		}

		csrfHeader := r.Header.Get("X-CSRF-TOKEN")
		if csrfHeader == "" {
			respondUnauthorized(w, "Cabeçalho CSRF não encontrado")
			return
		}

		// A validação crucial: o valor do cookie deve ser igual ao do cabeçalho
		if csrfCookie.Value != csrfHeader {
			respondUnauthorized(w, "Token CSRF inválido")
			return
		}
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

func validateToken(tokenString string) (*CustomClaims, error) {

	publicKeyBytes, err := os.ReadFile(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	if err != nil {
		return nil, errors.New("não foi possivel ler a chave pública")
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
	if err != nil {
		return nil, errors.New("não foi possivel fazer o parse da chave pública")
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("algoritmo de assinatura inesperado: %v", t.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err // O token é inválido (ex: assinatura errada)
	}

	// Agora, verifique se o token é válido E se as claims são do tipo correto
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token expirado")
		}
		return claims, nil
	}
	return nil, errors.New("token inválido")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
