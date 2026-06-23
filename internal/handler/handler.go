package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/db"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type Message struct {
	Data  string
	Type  string
	Where string
}

func HandleWS(w http.ResponseWriter, r *http.Request) {

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		messageJSON := string(p)
		fmt.Println(messageJSON)

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}

	// conn.WriteMessage(websocket.TextMessage, []byte("Websocket test"))
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RegisterResponse struct {
	Token string `json:"token"`
}

func HandleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()
	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, fmt.Sprintf("Bcrypt failed: %v", err), http.StatusBadRequest)
		return
	}
	var userID string
	err = db.Pool.QueryRow(ctx,
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id::text",
		req.Username, string(hash),
	).Scan(&userID)

	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID, "username": req.Username, "exp": time.Now().Add(24 * time.Hour).Unix()})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	data := RegisterResponse{
		Token: tokenString,
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
	Token string `json:"token"`
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx := r.Context()

	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to decode JSON: %v", err), http.StatusBadRequest)
		return
	}

	var passHash string
	var userID string

	err = db.Pool.QueryRow(ctx,
		"SELECT id, password_hash FROM users WHERE username = $1 LIMIT 1;",
		req.Username,
	).Scan(&userID, &passHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passHash), []byte(req.Password))
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID, "username": req.Username, "exp": time.Now().Add(24 * time.Hour).Unix()})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "failed to sign token", http.StatusInternalServerError)
		return
	}

	Login := RegisterResponse{
		Token: tokenString,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Login)
}

type ProfileResponse struct {
	Username string `json:"username"`
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		http.Error(w, "Invalid auth header format", http.StatusUnauthorized)
		return
	}

	tokenString := parts[1]

	token, err := jwt.ParseWithClaims(tokenString, &jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil

	})

	if err != nil || !token.Valid {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		http.Error(w, "invalid token claims", http.StatusUnauthorized)
		return
	}
	userID, ok := (*claims)["user_id"].(string)
	if !ok {
		http.Error(w, "invalid user_id claim", http.StatusUnauthorized)
		return
	}
	fmt.Println((*claims)["username"])
	username, ok := (*claims)["username"].(string)
	if !ok {
		http.Error(w, "invalid username claim", http.StatusUnauthorized)
		return
	}
	profile := ProfileResponse{
		Username: username,
		UserID:   userID,
		Token:    tokenString,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

