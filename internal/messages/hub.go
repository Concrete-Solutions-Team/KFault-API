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
	Broadcast  chan RoomMessage
	Register   chan *Sub
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

func NewHub(repo *Repository) *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		Broadcast:  make(chan RoomMessage, 1024),
		Register:   make(chan *Sub, 256),
		Unregister: make(chan *Client, 256),
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

			log.Println("Client data:", sub.Client)

			if newRoom == "" {
				if _, ok := h.Clients[client]; ok {
					log.Printf("Client already connected")
					succmsg := mustMarshal(Message{
						Type: TypeSystem,
						Payload: mustMarshal(SystemPayload{
							Message: "Client already connected",
							RoomID:  newRoom,
						}),
					})
					client.Send <- succmsg
				}

				h.Clients[client] = true
				log.Printf("Client connected first-time")
				succmsg := mustMarshal(Message{
					Type: TypeSystem,
					Payload: mustMarshal(SystemPayload{
						Message: "Client connected",
						RoomID:  newRoom,
					}),
				})
				client.Send <- succmsg
				continue
			}

			if client.RoomID == newRoom {
				log.Println("Already in this room, folks. No need to join again.")
				sysmsg := mustMarshal(SystemPayload{
					Message: "client already in the room",
					RoomID:  newRoom,
				})
				client.Send <- mustMarshal(Message{Type: TypeSystem, Payload: sysmsg})
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
			succmsg := mustMarshal(Message{
				Type: TypeSystem,
				Payload: mustMarshal(SystemPayload{
					Message: "Succesfully joined a room",
					RoomID:  newRoom,
				}),
			})
			client.Send <- succmsg

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
	client.Hub.Register <- &Sub{
		Client: client,
		RoomID: "",
	}

	go client.WritePump()
	go client.ReadPump()
}
