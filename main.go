package main

import (
	"fmt"
	"proxio/proxy"
	"proxio/ui"
)

func main() {
	var targetUrl = "http://localhost:8000"
	var proxyPortAccept = ":8001"
	var uiUrl = "http://127.0.0.1:4001"

	messagesChannel := proxy.ListenAndServe(proxyPortAccept, targetUrl)
	ui.Serve(uiUrl, messagesChannel)

	fmt.Printf("Forwarding: %s\t->\t%s\n", proxyPortAccept, targetUrl)
	fmt.Printf("Web interface: %s\n\n", uiUrl)

	select {}
}
