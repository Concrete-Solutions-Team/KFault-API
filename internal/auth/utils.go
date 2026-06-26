package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

func GetAuthInfo(r *http.Request) (*AuthInfo, error) {
	val := r.Context().Value(AuthContextKey)
	if val == nil {
		return nil, fmt.Errorf("no auth info in context")
	}

	return val.(*AuthInfo), nil
}

func (h *Handler) setAuthCookie(w http.ResponseWriter, token string) {
	sameSite := http.SameSiteLaxMode
	// check for secure http
	isSecure := strings.HasPrefix(h.frontendURL, "https://")
	if isSecure {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: sameSite,
		Path:     "/",
	})
}
