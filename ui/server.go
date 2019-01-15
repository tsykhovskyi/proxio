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

	ctr := NewController()

	mux := http.NewServeMux()
	mux.HandleFunc("/check", ctr.check)
	mux.HandleFunc("/clear", ctr.clear)
	mux.HandleFunc("/m", ctr.allMessages)
	mux.HandleFunc("/", ctr.static)

	go ctr.listenMessages(messagesChan)

	go func() {
		panic(http.ListenAndServe(localUrl.Host, mux))
	}()
}
