package server

import (
	"net/http"
	"strings"
)

func NewHttpServer(clientTrafficHandler, monitoringHandler http.Handler, monitoringDomain string) *http.Server {
	splitHandler := SubDomainMiddleware(
		clientTrafficHandler,
		monitoringHandler,
		monitoringDomain,
	)

	return &http.Server{
		Addr:    ":80",
		Handler: splitHandler,
	}
}

func SubDomainMiddleware(trafficHandler, monitorHandler http.Handler, monitoringDomain string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == monitoringDomain {
			monitorHandler.ServeHTTP(w, r)
			return
		}

		domainParts := strings.Split(r.Host, ".")
		if len(domainParts) == 2 {
			trafficHandler.ServeHTTP(w, r)
			return
		}

		PageNotFound(w)
	})
}
