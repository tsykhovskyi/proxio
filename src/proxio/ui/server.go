package ui

import (
	"github.com/gorilla/mux"
	"net/http"
	"proxio/client"
	"proxio/repository"
	"proxio/ssh"
)

func Handler(traffic client.Traffic, sessions repository.Sessions, balancer *ssh.Balancer) http.Handler {
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
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/test", ctr.test)
	api.HandleFunc("/clear", ctr.clear)
	api.HandleFunc("/m", ctr.domainTraffic)
	api.Handle("/ws", NewWsHandler(connectionPool))
	// r.PathPrefix("/").Handler(NewSpaHandler())
	r.PathPrefix("/").Handler(NewProxyHandler("http://localhost:4200"))

	api.Use(NewSessionMiddleware(sessions))
	api.Use(NewDomainPermissionMiddleware(balancer))

	return r
}
