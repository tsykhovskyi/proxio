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

	ctr := NewTrafficRequestHandler(storage)
	r := mux.NewRouter()
	r.HandleFunc("/clear", ctr.clear)
	r.HandleFunc("/m", ctr.domainTraffic)
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		connection := serveWs(w, r, connectionPool.closeChan)

		domain := r.URL.Query().Get("domain")
		if domain == "" {
			http.Error(w, "Domain not provided", 400)
		}
		connectionPool.NewConnection(domain, connection)
	})

	// r.PathPrefix("/").Handler(NewSpaHandler())
	proxyPath, _ := url.Parse("http://localhost:4200")
	proxy := httputil.NewSingleHostReverseProxy(proxyPath)
	r.PathPrefix("/").Handler(proxy)

	return r
}
