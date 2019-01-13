package ui

import (
	"net/http"
	"strconv"
)

func NewServer(port int, controller *Controller) *server {
	return &server{
		port: port,
		ctr:  controller,
	}
}

type server struct {
	port int
	ctr  *Controller
	srv  *http.Server
}

func (s server) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/check", s.ctr.check)
	mux.HandleFunc("/clear", s.ctr.clear)
	mux.HandleFunc("/m", s.ctr.allMessages)
	mux.HandleFunc("/", s.ctr.static)

	go func() {
		panic(http.ListenAndServe(":"+strconv.Itoa(s.port), mux))
	}()
}
