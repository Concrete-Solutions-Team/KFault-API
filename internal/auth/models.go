package auth

import "github.com/golang-jwt/jwt/v5"

type contextKey string

const (
	UserContextKey  contextKey = "user_claims"
	TokenContextKey contextKey = "raw_token"
	AuthContextKey  contextKey = "auth_key"
)

type CustomClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type AuthInfo struct {
	Claims CustomClaims
	Token  string
}
