package server

import (
	"log"
	"net/http"
	"strings"
)

func Start(configs *Configs) {
	balancer := NewBalancer()

	forwardingHandler := NewSshForwardHandler(balancer, configs.SshPort, configs.PrivateKeyPath)
	go func() {
		err := forwardingHandler.Start()
		log.Fatal(err)
	}()

	httpServer := &http.Server{
		Addr: ":80",
		Handler: middlewareSubdomain(
			balancer.httpHandler,
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		),
	}
	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	select {}
}

func middlewareSubdomain(subdomainHandler, originHandler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		domainParts := strings.Split(r.Host, ".")
		if len(domainParts) == 2 {
			subdomainHandler.ServeHTTP(w, r)
			return
		}

		originHandler.ServeHTTP(w, r)
	})
}
