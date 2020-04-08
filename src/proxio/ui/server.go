package ui

import (
	"net/http"
	"proxio/client"
)

func Serve(addr string, messagesChan chan *client.Message) http.Handler {
	connectionPool := NewConnectionPool()
	storage := NewStorage()

	go func() {
		for m := range messagesChan {
			storage.Add(m)
			connectionPool.BroadcastMessage(m.GetContext())
		}
	}()

	ctr := NewController(storage)
	mux := http.NewServeMux()
	mux.HandleFunc("/clear", ctr.clear)
	mux.HandleFunc("/m", ctr.allMessages)
	mux.HandleFunc("/", ctr.static)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connection := serveWs(w, r, connectionPool.closeChan)
		connectionPool.NewConnection(connection)
	})
	// go func() {
	// 	panic(http.ListenAndServe(addr, mux))
	// }()
	return mux
}
