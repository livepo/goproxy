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

		f, err := frame.DecodeFrame(msg)
		if err != nil {
			log.Println("Decode error:", err)
			continue
		}

		handleFrame(ws, f)
	}
}

func handleFrame(ws *websocket.Conn, f *frame.Frame) {
	switch f.Type {
	case 0x01: // 建立连接
		remote, err := connectRemote(f.ConnID, f.Data)
		if err != nil {
			log.Println("connect remote error:", err)
			ws.Close()
		}
		go sendToClient(remote, f.ConnID, ws)
	case 0x02: // 数据转发
		sendToRemote(f.ConnID, f.Data)

	case 0x03: // 心跳
		pong := &frame.Frame{Type: 0x03, ConnID: f.ConnID, Length: 0}
		ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(pong))
	}
}

func connectRemote(connID uint32, addr []byte) (remote net.Conn, err error) {
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
	return remote, err
}

// sendToClient 从remote接收数据发往客户端，注意不要阻塞发送通道
func sendToClient(remote net.Conn, connID uint32, ws *websocket.Conn) {
	buf := make([]byte, 4096)
	for {
		n, err := remote.Read(buf)
		if err != nil {
			remote.Close()
			connMap.Delete(connID)
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
}

func sendToRemote(connID uint32, data []byte) {
	if val, ok := connMap.Load(connID); ok {
		conn := val.(net.Conn)
		_, err := conn.Write(data)
		if err != nil {
			log.Println("send to remote error:", err)
			conn.Close()
			return
		}
	}
}
