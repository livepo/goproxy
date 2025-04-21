package main

import (
	socks5 "github.com/armon/go-socks5"
	"goproxy/internal/client"
	"log"
)

func main() {
	conf := &socks5.Config{
		Dial: client.DialThroughWebSocket, // 自定义 dialer
	}

	server, err := socks5.New(conf)
	if err != nil {
		log.Fatal("Socks5 server init failed:", err)
	}

	log.Println("SOCKS5 proxy listening on 127.0.0.1:1081")
	if err := server.ListenAndServe("tcp", "127.0.0.1:1081"); err != nil {
		log.Fatal("Socks5 Listen error:", err)
	}
}
