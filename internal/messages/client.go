package messages

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

// client -> hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- &Sub{
			Client: c,
			RoomID: c.RoomID,
		}
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))

	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

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
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Println("Ping failed, client disappeared. Kicking them out.")
				return
			}
		}
	}
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
