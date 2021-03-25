package ui

import (
	"github.com/gorilla/mux"
	"net/http"
	"proxio/ssh"
)

func Handler(balancer *ssh.Balancer) http.Handler {
	r := mux.NewRouter()
	r.Use(NewSessionMiddleware())

	api := r.PathPrefix("/api").Subrouter()

	ctr := NewTrafficRequestHandler()
	api.HandleFunc("/session", ctr.session)

	domain := api.PathPrefix("/domain").Subrouter()
	domain.Use(NewDomainPermissionMiddleware(balancer))

	domain.HandleFunc("/clear", ctr.clear)
	domain.HandleFunc("/m", ctr.domainTraffic)
	domain.Handle("/ws", NewWsHandler())

	r.PathPrefix("/").Handler(NewSpaHandler())
	// r.PathPrefix("/").Handler(NewProxyHandler("http://localhost:4200"))

	return r
}
