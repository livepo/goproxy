package main

import (
	"goproxy/internal/server"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ws", server.HandleWebSocket)

	addr := ":8080"
	log.Println("Proxy server listening on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
