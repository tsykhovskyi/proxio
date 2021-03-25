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
	api.HandleFunc("/session", ctr.session)

	domain := api.PathPrefix("/domain").Subrouter()
	domain.HandleFunc("/clear", ctr.clear)
	domain.HandleFunc("/m", ctr.domainTraffic)
	domain.Handle("/ws", NewWsHandler(connectionPool))
	r.PathPrefix("/").Handler(NewSpaHandler())
	// r.PathPrefix("/").Handler(NewProxyHandler("http://localhost:4200"))

	r.Use(NewSessionMiddleware(sessions))
	domain.Use(NewDomainPermissionMiddleware(balancer))

	return r
}
