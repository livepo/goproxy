package server

import (
	"github.com/gorilla/websocket"
	"goproxy/pkg/frame"
	"log"
	"net"
	"sync"
)

var connMap = sync.Map{} // ConnID → net.Conn

func handleClient(ws *websocket.Conn) {
	defer ws.Close()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			return
		}

		frame, err := frame.DecodeFrame(msg)
		if err != nil {
			log.Println("Decode error:", err)
			continue
		}

		handleFrame(ws, frame)
	}
}

func handleFrame(ws *websocket.Conn, f *frame.Frame) {
	switch f.Type {
	case 0x01: // 建立连接
		openRemote(f.ConnID, f.Data, ws)
	case 0x02: // 数据转发
		// time.Sleep(time.Second)
		if val, ok := connMap.Load(f.ConnID); ok {
			conn := val.(net.Conn)
			_, err := conn.Write(f.Data)
			if err != nil {
				log.Println("send to remote error:", err)
				return
			}
		}
	case 0x03: // 心跳
		pong := &frame.Frame{Type: 0x04, ConnID: f.ConnID, Length: 0}
		ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(pong))
	}
}

func openRemote(connID uint32, addr []byte, ws *websocket.Conn) {
	var err error
	var remote net.Conn
	conn, ok := connMap.Load(connID)
	if ok {
		remote = conn.(net.Conn)
	} else {
		target := string(addr) // e.g., "www.google.com:443"
		remote, err = net.Dial("tcp", target)
		if err != nil {
			log.Println("Dial error:", err)
			return
		}

		connMap.Store(connID, remote)
	}

	// 读取响应并返回给客户端
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := remote.Read(buf)
			if err != nil {
				if err.Error() != "EOF" {
					log.Println("Read from remote error:", err)
				}
				conn, ok := connMap.Load(connID)
				if ok {
					conn.(net.Conn).Close()
					connMap.Delete(connID)
				}
				return
			}

			resp := &frame.Frame{
				Type:   0x02,
				ConnID: connID,
				Length: uint32(n),
				Data:   buf[:n],
			}
			ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(resp))
		}
	}()
}
