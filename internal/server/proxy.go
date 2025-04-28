package server

import (
	"github.com/gorilla/websocket"
	"goproxy/pkg/frame"
	"io"
	"log"
	"net"
	"sync"
)

var connMap = sync.Map{} // ConnID → net.Conn

func handleClient(ws *websocket.Conn) {
	defer ws.Close()

	send := make(chan *frame.Frame)
	go sendFrame(ws, send)
	defer close(send)

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

		err = handleFrame(f, send)
		if err != nil {
			return
		}
	}
}

func sendFrame(ws *websocket.Conn, send chan *frame.Frame) {
	for {
		f := <-send
		ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(f))
	}
}

func handleFrame(f *frame.Frame, send chan *frame.Frame) error {
	switch f.Type {
	case 0x01: // 建立连接
		remote, err := connectRemote(f.ConnID, f.Data, send)
		if err != nil {
			log.Println("connect remote error:", err)
			return err
		}
		go sendToClient(remote, f.ConnID, send)
	case 0x02: // 数据转发
		sendToRemote(f.ConnID, f.Data)
	case 0x03: // 客户端主动断连
		conn, ok := connMap.Load(f.ConnID)
		if ok {
			remote := conn.(net.Conn)
			remote.Close()
			connMap.Delete(f.ConnID)
		}
	}
	return nil
}

func connectRemote(connID uint32, addr []byte, send chan *frame.Frame) (remote net.Conn, err error) {
	conn, ok := connMap.Load(connID)
	if ok {
		remote = conn.(net.Conn)
	} else {
		target := string(addr) // e.g., "www.google.com:443"
		remote, err = net.Dial("tcp", target)
		if err != nil {
			log.Println("Dial error:", err)
			closeClient(connID, send)
			return
		}

		connMap.Store(connID, remote)
	}
	return remote, err
}

// sendToClient 从remote接收数据发往客户端，注意不要阻塞发送通道
func sendToClient(remote net.Conn, connID uint32, send chan *frame.Frame) {
	buf := make([]byte, 4096)
	for {
		n, err := remote.Read(buf)
		if err != nil && err != io.EOF {
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
		send <- resp
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

func closeClient(connID uint32, send chan *frame.Frame) {
	f := &frame.Frame{
		Type:   0x03,
		ConnID: connID,
		Length: uint32(0),
	}
	send <- f
}
