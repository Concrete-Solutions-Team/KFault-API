package messages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type RegPayload struct {
	RoomID string `json:"room_id"`
}

type Sub struct {
	Client *Client
	RoomID string
}

func (h *Hub) Handle(msg Message, sender *Client) {
	switch msg.Type {
	case TypeChat:
		var payload ChatPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Println("bad chat payload", err)
			return
		}
		log.Printf("chat room_id: [%s]", payload.RoomID)
		fmt.Println(payload)
		msg2 := mustMarshal(ChatPayload{
			Text: payload.Text,
			Sender: sender.Auth.Claims.UserID,
			RoomID: payload.RoomID,
		})
		// msg.RoomID = msg.RoomID
		h.Broadcast <- Message{
			Type: TypeChat,
			Payload: msg2,
		}
	case TypeJoin:
		var payload RegPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Println("bad reg payload", err)
			return
		}
		log.Printf("join room_id: [%s]", payload.RoomID)

		h.Register <- &Sub{
			Client: sender,
			RoomID: payload.RoomID,
		}
	case TypeLeave:
		var payload RegPayload
		if err :=json.Unmarshal(msg.Payload, &payload); err != nil {
			log.Println("bad reg payload", err)
			return
		}
		log.Printf("leaving room_id: [%s]", payload.RoomID)
		h.Unregister <- &Sub{
			Client: sender,
			RoomID: payload.RoomID,
		}
	case TypePresence:
		log.Printf("sending presence request")
		h.Presence <- sender
	default:
		log.Println("unknown message type:", msg.Type)

	}
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
