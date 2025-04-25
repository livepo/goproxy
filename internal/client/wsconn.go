package client

import (
	"github.com/gorilla/websocket"
	"goproxy/internal/config"
	"goproxy/pkg/frame"
	"io"
	"net"
	"time"
)

type WSConn struct {
	ws     *websocket.Conn
	connID uint32
	rd     chan []byte
	closed chan struct{}
}

func NewWSConn(ws *websocket.Conn, id uint32) *WSConn {
	c := &WSConn{
		ws:     ws,
		connID: id,
		rd:     make(chan []byte, 1024),
		closed: make(chan struct{}),
	}

	go c.listen()
	return c
}

func (c *WSConn) Read(p []byte) (int, error) {
	select {
	case data := <-c.rd:
		n := copy(p, data)
		return n, nil
	case <-c.closed:
		return 0, io.EOF
	}
}

func (c *WSConn) Write(p []byte) (int, error) {
	f := &frame.Frame{
		Type:   0x02,
		ConnID: c.connID,
		Length: uint32(len(p)),
		Data:   p,
	}
	return len(p), c.ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(f))
}

func (c *WSConn) Close() error {
	return c.ws.Close()
}

func (c *WSConn) listen() {
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			close(c.closed)
			return
		}

		f, err := frame.DecodeFrame(msg)
		if err != nil || f.ConnID != c.connID {
			continue
		}

		c.rd <- f.Data
	}
}

func (c *WSConn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP(config.C.LocalHost),
		Port: config.C.LocalPort,
		Zone: "local",
	}
}
func (c *WSConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		Port: config.C.RemotePort,
		IP:   net.ParseIP(config.C.RemoteHost),
		Zone: "remote",
	}
}
func (c *WSConn) SetDeadline(t time.Time) error      { return nil }
func (c *WSConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *WSConn) SetWriteDeadline(t time.Time) error { return nil }
