package main

import (
	"flag"
	"fmt"
	"github.com/armon/go-socks5"
	"golang.org/x/net/context"
	"goproxy/internal/client"
	"goproxy/internal/config"
	"log"
	"net"
)

func main() {
	var configFile = flag.String("c", "config.yaml", "Path to config file")
	flag.Parse()

	config.MustLoad(*configFile)

	mux := client.NewMuxClient()

	go mux.StartReader()
	go mux.StartWriter()

	socks5Conf := &socks5.Config{
		// Dial: client.DialThroughWebSocket, // 自定义 dialer
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return mux.Dial(addr)
		},
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
