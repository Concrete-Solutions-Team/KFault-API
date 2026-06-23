package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/types"
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
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LoginResponse{Username: req.Username, Token: token})
}

type ProfileResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
}

func (h *Handler) Profile(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(types.UserContextKey)
	if val == nil {
		fmt.Println("Context value is nil!")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	claims, ok := val.(*types.CustomClaims)
	if !ok {
		fmt.Printf("Context value is of wrong type: %T\n", val)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	authHeader := r.Header.Get("Authorization")
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	profile := ProfileResponse{
		Username: claims.Username,
		UserID:   claims.UserID,
		Token:    tokenString,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}
