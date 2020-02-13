package ui

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"net"
	"net/http"
	"proxio/client"
)

type Connection struct {
	conn      net.Conn
	messages  chan *client.MessageContent
	closeChan chan bool
}

func (c *Connection) Send(m *client.MessageContent) error {
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

func serveWs(w http.ResponseWriter, r *http.Request, closeChannel chan *Connection) *Connection {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		fmt.Println("Server doesn't support ws")
	}

	connection := &Connection{conn, make(chan *client.MessageContent), make(chan bool, 1)}

	go func() {
		defer conn.Close()

		for {
			frame := ws.MustReadFrame(conn)
			if frame.Header.OpCode == ws.OpClose {
				closeChannel <- connection
				return
			}
		}
	}()
	return connection
}
