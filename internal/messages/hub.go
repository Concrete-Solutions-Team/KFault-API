package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/gorilla/websocket"
)

type Hub struct {
	// Clients    map[*Client]bool
	Rooms      map[string]map[*Client]bool
	Broadcast  chan RoomMessage
	Register   chan *Client
	Unregister chan *Client
	Repo       *Repository
}

type RoomMessage struct {
	RoomID  string `json:"room_id"`
	Payload []byte `json:"payload"`
}

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// todo: rooms implementation
func NewHub(repo *Repository) *Hub {
	return &Hub{
		Broadcast:  make(chan RoomMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Rooms:      make(map[string]map[*Client]bool),
		Repo:       repo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			if _, ok := h.Rooms[client.RoomID]; !ok {
				h.Rooms[client.RoomID] = make(map[*Client]bool)
			}
			log.Printf("registered client in room: %s, total: %d", client.RoomID, len(h.Rooms[client.RoomID]))
			h.Rooms[client.RoomID][client] = true
			log.Printf("Client joined room: %s", client.RoomID)
		case client := <-h.Unregister:
			if clients, ok := h.Rooms[client.RoomID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}

		case message := <-h.Broadcast:
			go func(roomID string, payload []byte) {
				var unm ChatPayload
				err := json.Unmarshal(payload, &unm)
				if err != nil {
					return
				}
				ctx := context.Background()
				if err = h.Repo.InsertMessage(ctx, MessageData{RoomID: roomID, Content: unm.Text, UserID: unm.Sender}); err != nil {
					log.Printf("Error saving message to DB for room %s: %v", roomID, err)
				}
			}(message.RoomID, message.Payload)
			log.Printf("broadcasting to room %s, clients: %d", message.RoomID, len(h.Rooms[message.RoomID]))
			if clients, ok := h.Rooms[message.RoomID]; ok {
				for client := range clients {
					select {
					case client.Send <- message.Payload:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}

			}
			// for client := range h.Rooms[message.RoomID] {
			// 	select {
			// 	case client.Send <- mustMarshal(message):
			// 	default:
			// 		close(client.Send)
			// 		delete(h.Rooms[message.RoomID], client)
			// 	}
			// }
		}
	}
}

func ServeWS(hub *Hub, db *Repository, w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	info, err := auth.GetAuthInfo(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on auth info parsing: %v", err), http.StatusBadRequest)
	}

	client := &Client{
		Hub:  hub,
		Conn: conn,
		Send: make(chan []byte, 256),
		Auth: info,
	}
	fmt.Println(info.Claims.UserID)
	client.Hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
