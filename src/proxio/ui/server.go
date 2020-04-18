package ui

import (
	"net/http"
	"proxio/client"
)

func Handler(publisher client.TrafficPublisher) http.Handler {
	connectionPool := NewConnectionPool()
	storage := NewStorage()

	go func() {
		for m := range publisher.GetTraffic() {
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
	return mux
}
