package handler

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
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
