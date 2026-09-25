package websocket

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, userID int64, userUUID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed for user %d: %v", userID, err)
		return
	}

	client := NewClient(hub, conn, userID, userUUID)
	hub.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()
}
