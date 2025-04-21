package client

import (
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"goproxy/pkg/frame"
	"log"
	"net"
	"net/http"
	"sync/atomic"
)

var connIDGen uint32 = 1

func DialThroughWebSocket(ctx context.Context, network, addr string) (net.Conn, error) {
	id := atomic.AddUint32(&connIDGen, 1)

	log.Println("Dialing via WebSocket:", addr)

	header := http.Header{}
	header.Set("X-Auth-Token", "your-secret")

	ws, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8080/ws", header)
	if err != nil {
		return nil, err
	}

	// 发送建立请求
	req := &frame.Frame{
		Type:   0x01,
		ConnID: id,
		Length: uint32(len(addr)),
		Data:   []byte(addr),
	}

	if err := ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(req)); err != nil {
		return nil, err
	}

	// 建立一个 WebSocketConn 结构实现 net.Conn 接口
	return NewWSConn(ws, id), nil
}
