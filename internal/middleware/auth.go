package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/golang-jwt/jwt/v5"
)

// TODO: refactor + utils

// internal/middleware/auth.go

func AuthMiddleware(repo *auth.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				if err == http.ErrNoCookie {
					http.Error(w, "Cookie not found", http.StatusNotFound)
					return
				}
				http.Error(w, "Server error", http.StatusInternalServerError)
				return
			}

			claims := &auth.CustomClaims{}
			token, err := jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(os.Getenv("JWT_SECRET")), nil
			})

			if err != nil || token == nil || !token.Valid {
				http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				return
			}

			isRevoked := repo.IsTokenRevoked(r.Context(), cookie.Value)
			fmt.Println(isRevoked)
			if isRevoked {
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}

			authData := auth.AuthInfo{
				Claims: *claims,
				Token:  cookie.Value,
			}

			ctx := context.WithValue(r.Context(), auth.AuthContextKey, authData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
