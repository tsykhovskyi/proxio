package ui

import (
	"fmt"
	"net/http"
	"net/url"
	"proxio/proxy"
)

func Serve(local string, messagesChan chan *proxy.Message) {
	localUrl, err := url.Parse(local)
	if err != nil {
		panic(fmt.Sprintf("Unable to parse url: %s", err))
	}

	connectionPool := NewConnectionPool()
	storage := NewStorage()

	go func() {
		for m := range messagesChan {
			storage.Add(m)
			connectionPool.BroadcastMessage(m.GetContext())
			fmt.Println("New message", len(connectionPool.Connections))
		}
	}()

	ctr := NewController(storage)
	mux := http.NewServeMux()
	mux.HandleFunc("/clear", ctr.clear)
	mux.HandleFunc("/m", ctr.allMessages)
	mux.HandleFunc("/", ctr.static)
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn := serveWs(w, r)
		connectionPool.NewConnection(conn)
		fmt.Println("New connection", len(connectionPool.Connections))
	})
	go func() {
		panic(http.ListenAndServe(localUrl.Host, mux))
	}()
}
