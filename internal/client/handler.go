package client

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"goproxy/internal/config"
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
	header.Set("X-Auth-Token", config.C.Password)

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // true 用于自签名证书调试，正式必须 false
		},
	}

	websocketURL := fmt.Sprintf("wss://%s:%d/ws", config.C.RemoteHost, config.C.RemotePort)
	ws, _, err := dialer.Dial(websocketURL, header)
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
