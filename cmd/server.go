package main

import (
	"flag"
	"fmt"
	"goproxy/internal/config"
	"goproxy/internal/server"
	"log"
	"net/http"
)

func main() {
	var configFile = flag.String("c", "config.yaml", "Path to config file")
	flag.Parse()

	config.MustLoad(*configFile)

	http.HandleFunc("/ws", server.HandleWebSocket)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello, world"))
	})

	log.Println("Proxy server listening on", config.C.RemotePort)
	err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", config.C.RemoteHost, config.C.RemotePort), config.C.CertPath, config.C.KeyPath, nil)
	if err != nil {
		log.Fatal("Server failed:", err)
	}
}
