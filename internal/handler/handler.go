package handler

// type Message struct {
// 	Data  string
// 	Type  string
// 	Where string
// }

// func HandleWS(w http.ResponseWriter, r *http.Request) {

// 	var upgrader = websocket.Upgrader{
// 		ReadBufferSize:  1024,
// 		WriteBufferSize: 1024,
// 		CheckOrigin: func(r *http.Request) bool {
// 			return true
// 		},
// 	}

// 	conn, err := upgrader.Upgrade(w, r, nil)

// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	for {
// 		messageType, p, err := conn.ReadMessage()
// 		if err != nil {
// 			log.Println(err)
// 			return
// 		}
// 		messageJSON := string(p)
// 		fmt.Println(messageJSON)

// 		if err := conn.WriteMessage(messageType, p); err != nil {
// 			log.Println(err)
// 			return
// 		}

// 	}

// 	// conn.WriteMessage(websocket.TextMessage, []byte("Websocket test"))
// }

// type RegisterRequest struct {
// 	Username string `json:"username"`
// 	Password string `json:"password"`
// }
// type RegisterResponse struct {
// 	Token string `json:"token"`
// }

// func HandleRegister(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// 	ctx := r.Context()
// 	var req RegisterRequest

// 	err := json.NewDecoder(r.Body).Decode(&req)
// 	if err != nil {
// 		http.Error(w, "invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Bcrypt failed: %v", err), http.StatusBadRequest)
// 		return
// 	}
// 	var userID string
// 	err = db.Pool.QueryRow(ctx,
// 		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id::text",
// 		req.Username, string(hash),
// 	).Scan(&userID)

// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Query failed: %v", err), http.StatusBadRequest)
// 		return
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"user_id": userID, "exp": time.Now().Add(24 * time.Hour).Unix()})

// 	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
// 	if err != nil {
// 		http.Error(w, "failed to sign token", http.StatusInternalServerError)
// 		return
// 	}

// 	w.WriteHeader(http.StatusOK)
// 	data := RegisterResponse{
// 		Token: tokenString,
// 	}
// 	if err = json.NewEncoder(w).Encode(data); err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to encode JSON: %v", err), http.StatusBadRequest)
// 		return
// 	}
// }
