package ui

import (
	"fmt"
	gows "github.com/gobwas/ws"
	"net/http"
	"proxio/ws"
)

func serveWs(w http.ResponseWriter, r *http.Request, closeChannel chan *ws.Connection) *ws.Connection {
	conn, _, _, err := gows.UpgradeHTTP(r, w)
	if err != nil {
		fmt.Println("Server doesn't support ws")
	}

	connection := ws.NewConnection(conn)

	go func() {
		defer conn.Close()

		for {
			frame, err := gows.ReadFrame(conn)
			if err != nil || frame.Header.OpCode == gows.OpClose {
				closeChannel <- connection
				return
			}
		}
	}()

	return connection
}
