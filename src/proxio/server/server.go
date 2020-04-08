package server

import (
	"log"
	"net/http"
	"proxio/client"
	"proxio/ui"
	"strings"
)

func Start(configs *Configs) {
	balancer := NewBalancer()

	forwardingHandler := NewSshForwardHandler(balancer, configs.SshPort, configs.PrivateKeyPath)
	go func() {
		err := forwardingHandler.Start()
		log.Fatal(err)
	}()

	messagesChan := make(chan *client.Message, 1)
	trackedSubDomainBalancer := client.TrafficMiddleware(messagesChan, balancer.httpHandler)
	ui.Serve(":4000", messagesChan)

	splitHandler := SubDomainMiddleware(
		trackedSubDomainBalancer,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("<h1>Forward path not found</h1>"))
		}),
	)

	httpServer := &http.Server{
		Addr:    ":80",
		Handler: splitHandler,
	}
	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	select {}
}

func SubDomainMiddleware(subDomainHandler, originHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domainParts := strings.Split(r.Host, ".")
		if len(domainParts) == 2 {
			subDomainHandler.ServeHTTP(w, r)
			return
		}

		originHandler.ServeHTTP(w, r)
	})
}
