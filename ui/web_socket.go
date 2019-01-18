package ui

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"net"
	"net/http"
	"proxio/proxy"
)

type Connection struct {
	conn     net.Conn
	messages chan *proxy.MessageContent
}

func (c *Connection) Send(m *proxy.MessageContent) error {
	payload, err := json.Marshal(m)
	if err != nil {
		panic("unable to encode frame")
	}
	frame := ws.NewTextFrame(payload)

	if err = ws.WriteFrame(c.conn, frame); err != nil {
		return err
	}
	return nil
}

func serveWs(w http.ResponseWriter, r *http.Request) *Connection {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		fmt.Println("Server doesn't support ws")
	}

	go func() {
		defer conn.Close()

		for {
			frame := ws.MustReadFrame(conn)
			if frame.Header.OpCode == ws.OpClose {
				return
			}
		}
	}()
	return &Connection{conn, make(chan *proxy.MessageContent)}
}
