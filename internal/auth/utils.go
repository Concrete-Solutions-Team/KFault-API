package auth

import (
	"fmt"
	"net/http"
	"time"
)

func GetAuthInfo(r *http.Request) (*AuthInfo, error) {
	val := r.Context().Value(AuthContextKey)
	if val == nil {
		return nil, fmt.Errorf("no auth info in context")
	}

	return val.(*AuthInfo), nil
}

func SetAuthCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}
