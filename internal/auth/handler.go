package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/helpers"
)

type Message struct {
	Data  string
	Type  string
	Where string
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type AuthResponse struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req AuthRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	token, err := h.service.Register(ctx, &UserReq{Username: req.Username, Password: req.Password})
	if err != nil {
		log.Printf("register failed: %v", err)
		http.Error(w, fmt.Sprintf("Registration failed: %v", err), http.StatusBadRequest)
		return
	}

	SetAuthCookie(w, token)

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(&AuthResponse{
		Username: req.Username,
		Message:  "Registration successful",
	}); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusBadRequest)
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), &UserReq{Username: req.Username, Password: req.Password})
	if err != nil {
		log.Printf("Login failed for user %s: %v", req.Username, err)
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	SetAuthCookie(w, token)

	helpers.SendJSON(w, http.StatusOK, &AuthResponse{
		Username: req.Username,
		Message:  "Login successful",
	})
}

type ProfileResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	authInfo, err := GetAuthInfo(r)
	if err != nil || authInfo == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile := ProfileResponse{
		Username: authInfo.Claims.Username,
		UserID:   authInfo.Claims.UserID,
		Token:    authInfo.Token,
	}

	helpers.SendJSON(w, http.StatusOK, profile)
}

func (h *Handler) LogOut(w http.ResponseWriter, r *http.Request) {
	authInfo, err := GetAuthInfo(r)
	if err != nil || authInfo == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()

	h.service.LogOut(ctx, authInfo)
	profile := ProfileResponse{
		Token: authInfo.Token,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	helpers.SendJSON(w, http.StatusOK, profile)
}
