package main

import (
	"fmt"
	"proxio/proxy"
	"proxio/ui"
	"strconv"
)

func main() {
	target := "http://localhost:8000"
	var local int = 8001
	var uiPort int = 4001

	stor := proxy.NewStorage()
	p := proxy.NewProxy(local, target, stor)
	go p.Serve()

	ctr := &ui.Controller{p.Messages, p.Storage}
	server := ui.NewServer(uiPort, ctr)
	server.Serve()

	fmt.Printf("Forwarding: %s\t->\t%s\n", "localhost:"+strconv.Itoa(local), target)
	fmt.Printf("Web interface: %s\n\n", "http://localhost:"+strconv.Itoa(uiPort))

	select {}
}

func listenMessages(messages chan *proxy.Message) {
	for m := range messages {
		if m.HasResponse() {
			fmt.Printf("%d:\t[%s]\t%-20s%d\n", m.Id, m.Request.Method, m.Request.RequestURI, m.Response.StatusCode)
		} else {
			fmt.Printf("%d:\t[%s]\t%-20s\n", m.Id, m.Request.Method, m.Request.RequestURI)
		}
	}
}
