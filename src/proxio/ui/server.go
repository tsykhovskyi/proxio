package ui

import (
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httputil"
	"net/url"
	"proxio/client"
)

func Handler(traffic client.Traffic) http.Handler {
	connectionPool := NewConnectionPool()
	storage := NewStorage()

	go func() {
		for m := range traffic {
			storage.Add(m)
			connectionPool.BroadcastMessage(m.GetContext())
		}
	}()

	ctr := NewController(storage)
	r := mux.NewRouter()
	r.HandleFunc("/clear", ctr.clear)
	r.HandleFunc("/m", ctr.allMessages)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connection := serveWs(w, r, connectionPool.closeChan)
		connectionPool.NewConnection(connection)
	})

	// r.PathPrefix("/").Handler(NewSpaHandler())
	proxyPath, _ := url.Parse("http://localhost:4200")
	proxy := httputil.NewSingleHostReverseProxy(proxyPath)
	r.PathPrefix("/").Handler(proxy)

	return r
}
