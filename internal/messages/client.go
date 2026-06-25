package messages

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// client -> hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("raw received: %s", raw)
		var msg Message
		if err := json.Unmarshal(raw, &msg); err != nil {
			log.Println("bad message", err)
			continue
		}
		log.Printf("parsed type: %s", msg.Type)
		c.Hub.Handle(msg, c)
		// c.Hub.Broadcast <- raw
	}
}

// hub -> websocket
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
		log.Println(message)
	}

	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) Emit(msgType MessageType, payload any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := Message{Type: msgType, Payload: raw}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	c.Send <- data
	return nil
}
