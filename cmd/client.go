package main

import (
	"flag"
	"fmt"
	"github.com/armon/go-socks5"
	"goproxy/internal/client"
	"goproxy/internal/config"
	"log"
)

func main() {
	var configFile = flag.String("c", "config.yaml", "Path to config file")
	flag.Parse()

	config.MustLoad(*configFile)

	socks5Conf := &socks5.Config{
		Dial: client.DialThroughWebSocket, // 自定义 dialer
	}

	server, err := socks5.New(socks5Conf)
	if err != nil {
		log.Fatal("Socks5 server init failed:", err)
	}

	addr := fmt.Sprintf("%s:%d", config.C.LocalHost, config.C.LocalPort)
	log.Println("SOCKS5 proxy listening on ", addr)
	if err := server.ListenAndServe("tcp", addr); err != nil {
		log.Fatal("Socks5 Listen error:", err)
	}
}
