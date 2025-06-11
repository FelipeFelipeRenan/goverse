package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/FelipeFelipeRenan/goverse/api-gateway/pkg/jwtutils"
)

type authContextKey string

const UserIDKey authContextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromHeader(r)
		if token == "" {
			http.Error(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		claims, err := jwtutils.ValidateToken(token)
		if err != nil {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		// Use diretamente claims.UserID
		userID := claims.UserID

		fmt.Println("AuthMiddleware userID:", userID)

		if userID == "" {
			http.Error(w, "Token inválido (sem user_id)", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}
