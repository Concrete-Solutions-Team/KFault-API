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

// hub is the boss
type Hub struct {
	Clients    map[*Client]bool
	Rooms      map[string]map[*Client]bool
	Broadcast  chan Message
	Register   chan *Sub
	Unregister chan *Sub
	Presence   chan *Client
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

func NewHub(repo *Repository) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan Message, 1024),
		Register:   make(chan *Sub, 256),
		Unregister: make(chan *Sub, 256),
		Presence:   make(chan *Client, 256),
		Rooms:      make(map[string]map[*Client]bool),
		Repo:       repo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.Register:
			client := sub.Client
			newRoom := sub.RoomID

			if _, exists := h.Rooms[newRoom][client]; exists {
				log.Println("Look, folks, this user is already in the room. They are already there! We love our users, but they don't need to join twice. It is redundant, okay? No need to join again. Tremendous.")
				syspd := mustMarshal(SystemPayload{
					Message: "Let me tell you, folks, an instance of this user is already in this room, folks. No need to join again. Tremendous.",
					RoomID:  newRoom,
				})
				sysmsg := mustMarshal(Message{
					Type:    TypeSystem,
					Payload: syspd,
				})
				client.Send <- sysmsg
				continue
			}

			if newRoom == "" {
				if _, ok := h.Clients[client]; ok {
					log.Printf("Client already connected")
					syspd := mustMarshal(SystemPayload{
						Message: "Client already connected",
						RoomID:  newRoom,
					})
					msg := mustMarshal(Message{
						Type:    TypeSystem,
						Payload: syspd,
					})
					client.Send <- msg
				}

				h.Clients[client] = true
				log.Printf("Client connected first-time")
				syspd := mustMarshal(SystemPayload{
					Message: "Client connected",
					RoomID:  newRoom,
				})
				succmsg := mustMarshal(Message{
					Type:    TypeSystem,
					Payload: syspd,
				})
				client.Send <- succmsg
				continue
			}

			if client.RoomID == newRoom {
				log.Println("Already in this room, folks. No need to join again.")
				syspd := mustMarshal(SystemPayload{
					Message: "client already in the room",
					RoomID:  newRoom,
				})
				sysmsg := mustMarshal(Message{Type: TypeSystem, Payload: syspd})
				client.Send <- sysmsg
				continue
			}

			if client.RoomID != "" {
				if oldRoomClients, ok := h.Rooms[client.RoomID]; ok {
					delete(oldRoomClients, client)
					log.Printf("Removed client from old room: %s", client.RoomID)

					if len(oldRoomClients) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}
			if _, ok := h.Rooms[newRoom]; !ok {
				h.Rooms[newRoom] = make(map[*Client]bool)
			}
			h.Rooms[newRoom][client] = true

			client.RoomID = newRoom
			log.Printf("Client successfully joined new room: %s", newRoom)
			succpd := mustMarshal(SystemPayload{
				Message: "Succesfully joined a room",
				RoomID:  newRoom,
			})
			succmsg := mustMarshal(Message{
				Type:    TypeSystem,
				Payload: succpd,
			})
			client.Send <- succmsg
			if clients, ok := h.Rooms[newRoom]; ok {
				var list []ClientInfo
				if room, exists := h.Rooms[client.RoomID]; exists {
					for c := range room {
						if c.Auth != nil {
							list = append(list, ClientInfo{Username: c.Auth.Claims.Username})
						}
					}
				}
				bytes := mustMarshal(list)
				presence := mustMarshal(Message{
					Type:    TypePresence,
					Payload: bytes,
				})
				for cli := range clients {
					cli.Send <- presence
				}
			}

		case sub := <-h.Unregister:
			client := sub.Client
			room := sub.RoomID
			if _, exists := h.Clients[client]; exists {
				delete(h.Clients, client)
				log.Printf("Cleaned client from global registry.")
			}
			if room != "" {
				if clients, ok := h.Rooms[room]; ok {
					if _, ok := clients[client]; ok {
						delete(clients, client)
						log.Printf("Removed client from room: %s", room)
					}

					if len(clients) == 0 {
						delete(h.Rooms, room)
						log.Printf("Room %s is completely empty. Room deleted.", room)
					}
				}
			}
			if client.Conn != nil {
				client.Conn.Close()
				log.Println("Closed raw WebSocket connection network pipe safely.")
			}

		case message := <-h.Broadcast:
			var unm ChatPayload
			err := json.Unmarshal(message.Payload, &unm)
			if err != nil {
				continue
			}
			go func(roomID string) {
				ctx := context.Background()
				log.Println("async run")
				if err = h.Repo.InsertMessage(ctx, MessageData{RoomID: roomID, Content: unm.Text, UserID: unm.Sender}); err != nil {
					log.Printf("Error saving message to DB for room %s: %v", roomID, err)
				}
			}(unm.RoomID)

			log.Printf("broadcasting to room %s, clients: %d", unm.RoomID, len(h.Rooms[unm.RoomID]))
			payload := mustMarshal(unm)

			msg := mustMarshal(Message{
				Type:    TypeChat,
				Payload: payload,
			})

			if clients, ok := h.Rooms[unm.RoomID]; ok {
				for client := range clients {
					select {
					case client.Send <- msg:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
		case client := <-h.Presence:
			log.Printf("client requested presence: %v", client.Auth.Claims.UserID)
			if room, exists := h.Rooms[client.RoomID]; exists {
				var list []ClientInfo
				for c := range room {
					if c.Auth != nil {
						list = append(list, ClientInfo{Username: c.Auth.Claims.Username})
					}
				}

				listBytes := mustMarshal(list)
				msg := mustMarshal(Message{
					Type:    TypePresence,
					Payload: listBytes,
				})

				client.Send <- msg
			}
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
	client.Hub.Register <- &Sub{
		Client: client,
		RoomID: "",
	}

	go client.WritePump()
	go client.ReadPump()
}
