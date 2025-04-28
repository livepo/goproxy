package client

import (
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"goproxy/internal/config"
	"goproxy/pkg/frame"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type MuxClient struct {
	ws    *websocket.Conn
	conns map[uint32]*MuxStream
	coMu  sync.RWMutex
	send  chan *frame.Frame
}

func NewMuxClient() *MuxClient {
	ws, err := DialWebsocket()

	if err != nil {
		log.Fatal(err)
	}

	m := MuxClient{
		ws:    ws,
		conns: make(map[uint32]*MuxStream),
		coMu:  sync.RWMutex{},
		send:  make(chan *frame.Frame, 1024),
	}

	return &m
}

func DialWebsocket() (ws *websocket.Conn, err error) {
	header := http.Header{}
	header.Set("X-Auth-Token", config.C.Password)

	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // true 用于自签名证书调试，正式必须 false
		},
		EnableCompression: true,
	}

	websocketURL := fmt.Sprintf("wss://%s:%d/ws", config.C.RemoteHost, config.C.RemotePort)
	ws, _, err = dialer.Dial(websocketURL, header)

	return ws, err
}

func (c *MuxClient) Dial(addr string) (net.Conn, error) {
	connID := atomic.AddUint32(&connIDGen, 1)
	stream := &MuxStream{
		connID: connID,
		client: c,
		rbuf:   make(chan []byte, 4096),
	}

	c.coMu.Lock()
	c.conns[connID] = stream
	c.coMu.Unlock()

	f := &frame.Frame{Type: 0x01, ConnID: connID, Length: uint32(len(addr)), Data: []byte(addr)}
	_ = c.sendFrame(f)

	return stream, nil
}

func (c *MuxClient) sendFrame(f *frame.Frame) error {
	fmt.Println("sendFrame...: f.Type, f.ConnID", f.Type, f.ConnID)
	c.send <- f
	fmt.Println("send through send channel finished...f.Type, f.ConnID", f.Type, f.ConnID)
	return nil
}

func (c *MuxClient) removeConn(connID uint32) {
	c.coMu.Lock()
	delete(c.conns, connID)
	c.coMu.Unlock()
}

func (c *MuxClient) StartReader() {
	fmt.Println("start read websocket msg to socks5 tunnel...")
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		f, err := frame.DecodeFrame(msg)
		if err != nil {
			continue
		}
		c.coMu.RLock()
		stream, ok := c.conns[f.ConnID]
		c.coMu.RUnlock()
		if ok && f.Type == 0x02 {
			stream.rbuf <- f.Data
		} else if f.Type == 0x03 {
			close(stream.rbuf)
			c.removeConn(f.ConnID)
		}
	}
}

func (c *MuxClient) StartWriter() {
	fmt.Println("start write data to websocket server tunnel...")
	for {
		fmt.Println("read from send channel begin")
		f := <-c.send
		fmt.Println("read from send channel finished: f.Type, f.ConnID", f.Type, f.ConnID)

		for {
			err := c.ws.WriteMessage(websocket.BinaryMessage, frame.EncodeFrame(f))
			if err != nil {
				log.Println("write to websocket server error, ", err, "reconnecting")
				c.ws, err = DialWebsocket()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				break
			}
		}
	}
}

type MuxStream struct {
	connID uint32
	client *MuxClient
	rbuf   chan []byte
}

// Read socks5 invoke Read to read data
func (s *MuxStream) Read(p []byte) (int, error) {
	data, ok := <-s.rbuf
	if !ok {
		return 0, io.EOF
	}
	return copy(p, data), nil
}

// Write socks5 invoke Write method
func (s *MuxStream) Write(p []byte) (int, error) {
	f := &frame.Frame{Type: 0x02, ConnID: s.connID, Length: uint32(len(p)), Data: p}
	fmt.Println("socks5 write ...", f.Type, f.ConnID)
	return len(p), s.client.sendFrame(f)
}

func (s *MuxStream) Close() error {
	fmt.Println("socks5 close....")
	f := &frame.Frame{Type: 0x03, ConnID: s.connID}
	s.client.sendFrame(f)
	s.client.removeConn(s.connID)
	return nil
}

func (s *MuxStream) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.ParseIP(config.C.LocalHost),
		Port: config.C.LocalPort,
		Zone: "local",
	}
}

func (s *MuxStream) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		Port: config.C.RemotePort,
		IP:   net.ParseIP(config.C.RemoteHost),
		Zone: "remote",
	}
}

func (s *MuxStream) SetDeadline(t time.Time) error {
	return nil
}

func (s *MuxStream) SetReadDeadline(t time.Time) error {
	return nil
}

func (s *MuxStream) SetWriteDeadline(t time.Time) error {
	return nil
}
