package auth

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const (
	UserContextKey  contextKey = "user_claims"
	TokenContextKey contextKey = "raw_token"
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
