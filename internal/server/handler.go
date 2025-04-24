package server

import (
	"goproxy/internal/config"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 可根据需求限制来源
	},
	EnableCompression: true,
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 鉴权（可选）
	if r.Header.Get("X-Auth-Token") != config.C.Password {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	log.Println("New WebSocket connection from", r.RemoteAddr)

	handleClient(conn)
}
