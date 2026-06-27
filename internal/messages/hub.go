package messages

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/Concrete-Solutions-Team/KFault-API/internal/rooms"
	"github.com/gorilla/websocket"
)

// hub is the boss

type Room struct {
	Clients map[*Client]bool
	History []ChatPayload
}

type Hub struct {
	Clients      map[*Client]bool
	Rooms        map[string]*Room
	Broadcast    chan Message
	Register     chan *Sub
	Unregister   chan *Sub
	Presence     chan *Client
	Repo         *Repository
	RoomsService *rooms.Service
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

func NewHub(repo *Repository, rs *rooms.Service) *Hub {
	return &Hub{
		Clients:      make(map[*Client]bool),
		Broadcast:    make(chan Message, 1024),
		Register:     make(chan *Sub, 256),
		Unregister:   make(chan *Sub, 256),
		Presence:     make(chan *Client, 256),
		Rooms:        make(map[string]*Room),
		Repo:         repo,
		RoomsService: rs,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.Register:
			client := sub.Client
			newRoom := sub.RoomID

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

			room, ok := h.Rooms[newRoom]
			if !ok {
				if _, err := h.RoomsService.GetRoomByID(context.Background(), newRoom); err != nil {
					log.Println("Room doesn't exist")
					syspd := mustMarshal(SystemPayload{
						Message: "Let me tell you, folks, room instance doesnt exist. Tremendous.",
						RoomID:  newRoom,
					})
					sysmsg := mustMarshal(Message{
						Type:    TypeSystem,
						Payload: syspd,
					})
					client.Send <- sysmsg
					continue
				}

				his, err := h.Repo.GetMessagesByRoom(context.Background(), newRoom, 50)
				if err != nil {
					log.Printf("Error on recovering history: %v", err)
				}
				log.Println("history: ", his)
				h.Rooms[newRoom] = &Room{
					Clients: make(map[*Client]bool),
					History: his,
				}

				h.Rooms[newRoom].Clients[client] = true
				room = h.Rooms[newRoom]

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

				hispd := mustMarshal(HistoryPayload{
					Messages: room.History,
				})
				hismsg := mustMarshal(Message{
					Type:    TypeHistory,
					Payload: hispd,
				})
				client.Send <- hismsg
				continue
			}

			if room.Clients[client] {
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
				if room, ok := h.Rooms[client.RoomID]; ok {
					oldRoomClients := room.Clients
					delete(oldRoomClients, client)
					log.Printf("Removed client from old room: %s", client.RoomID)

					if len(oldRoomClients) == 0 {
						delete(h.Rooms, client.RoomID)
					}
				}
			}

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
			hispd := mustMarshal(HistoryPayload{
				Messages: room.History,
			})
			hismsg := mustMarshal(Message{
				Type:    TypeHistory,
				Payload: hispd,
			})
			client.Send <- hismsg
			if ok {
				var list []ClientInfo
				for c := range room.Clients {
					if c.Auth != nil {
						list = append(list, ClientInfo{Username: c.Auth.Claims.Username})
					}
				}

				bytes := mustMarshal(list)
				presence := mustMarshal(Message{
					Type:    TypePresence,
					Payload: bytes,
				})
				for cli := range room.Clients {
					cli.Send <- presence
				}
			}

		case sub := <-h.Unregister:
			client := sub.Client
			roomToRemove := sub.RoomID
			if _, exists := h.Clients[client]; exists {
				delete(h.Clients, client)
				log.Printf("Cleaned client from global registry.")
			}
			if roomToRemove != "" {
				if room, ok := h.Rooms[roomToRemove]; ok {
					if _, ok := room.Clients[client]; ok {
						delete(room.Clients, client)
						log.Printf("Removed client from room: %s", roomToRemove)
					}

					if len(room.Clients) == 0 {
						delete(h.Rooms, roomToRemove)
						log.Printf("Room %s is completely empty. Room deleted.", roomToRemove)
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
			room, ok := h.Rooms[unm.RoomID]
			if !ok || room == nil {
				continue
			}
			go func(roomID string) {
				ctx := context.Background()
				log.Println("async run")
				if dberr := h.Repo.InsertMessage(ctx, MessageData{RoomID: roomID, Content: unm.Text, UserID: unm.Sender}); dberr != nil {
					log.Printf("Error saving message to DB for room %s: %v", roomID, dberr)
				}
			}(unm.RoomID)

			log.Printf("broadcasting to room %s, clients: %d", unm.RoomID, len(h.Rooms[unm.RoomID].Clients))
			payload := mustMarshal(unm)
			h.Rooms[unm.RoomID].History = append(h.Rooms[unm.RoomID].History, unm)
			msg := mustMarshal(Message{
				Type:    TypeChat,
				Payload: payload,
			})

			if room, ok := h.Rooms[unm.RoomID]; ok {
				for client := range room.Clients {
					select {
					case client.Send <- msg:
					default:
						close(client.Send)
						delete(room.Clients, client)
					}
				}
			}
		case client := <-h.Presence:
			log.Printf("client requested presence: %v", client.Auth.Claims.UserID)
			if room, exists := h.Rooms[client.RoomID]; exists {
				var list []ClientInfo
				for c := range room.Clients {
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
