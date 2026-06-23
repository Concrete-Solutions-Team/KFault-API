package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Message struct {
	Data  string
	Type  string
	Where string
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	token, err := h.service.Register(ctx, &UserReq{Username: req.Username, Password: req.Password})
	if err != nil {
		http.Error(w, fmt.Sprintf("Register failed: %v", err), http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
	data := RegisterResponse{
		Username: req.Username,
		Token:    token,
	}
	if err = json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusBadRequest)
		return
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode JSON: %v", err), http.StatusBadRequest)
		return
	}
	token, err := h.service.Login(ctx, &UserReq{Username: req.Username, Password: req.Password})
	if err != nil {
		http.Error(w, fmt.Sprintf("internal server error: %v", err), http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Username: req.Username, Token: token})
}

type ProfileResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(UserContextKey).(*CustomClaims)
	tokenString := r.Context().Value(TokenContextKey).(string)

	profile := ProfileResponse{
		Username: claims.Username,
		UserID:   claims.UserID,
		Token:    tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

func (h *Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tokenString := ctx.Value(TokenContextKey).(string)

	h.service.LogOut(ctx, tokenString)
	// h.service.repo.ExpireToken(r.Context(), tokenString, *claims)
	profile := ProfileResponse{
		Token:    tokenString,
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}
