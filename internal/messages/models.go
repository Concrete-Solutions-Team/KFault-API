package messages

import (
	"encoding/json"

	"github.com/Concrete-Solutions-Team/KFault-API/internal/auth"
	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	RoomID string `json:"room_id"`
	Send   chan []byte
	Auth *auth.AuthInfo
}

const (
	TypeChat    MessageType = "chat"
	TypeSystem  MessageType = "system"
	TypeJoin    MessageType = "join"
	TypeLeave   MessageType = "leave"
	TypeHistory MessageType = "history"
)

type MessageType string

type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
	RoomID  string          `json:"room_id"`
}

type ChatPayload struct {
	Text   string `json:"text"`
	Sender string `json:"user_id"`
	RoomID string `json:"room_id"`
}

type JoinPayload struct {
	RoomID string `json:"room_id"`
	Sender string `json:"sender"`
}

type HistoryPayload struct {
	Messages []StoredMessage `json:"messages"`
}

type StoredMessage struct {
	ID     string `json:"id"`
	RoomID string `json:"room_id"`
	Sender string `json:"sender"`
	Text   string `json:"text"`
	SentAt string `json:"sent_at"`
}

type SystemPayload struct {
	Text string `json:"text"`
}

func mustMarshal(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
