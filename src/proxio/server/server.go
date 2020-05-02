package server

import (
	"net/http"
)

func NewHttpServer(clientTrafficHandler, monitoringHandler http.Handler, host, monitoringDomain string) *http.Server {
	splitHandler := SubDomainMiddleware(
		clientTrafficHandler,
		monitoringHandler,
		host,
		monitoringDomain,
	)

	return &http.Server{
		Addr:    ":80",
		Handler: splitHandler,
	}
}

func SubDomainMiddleware(trafficHandler, monitorHandler http.Handler, host, monitoringDomain string) http.Handler {
	// subdomainPartsSize := len(strings.Split(host, "."))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host == monitoringDomain {
			monitorHandler.ServeHTTP(w, r)
			return
		}

		// domainParts := strings.Split(r.Host, ".")
		// if len(domainParts) == subdomainPartsSize {
		trafficHandler.ServeHTTP(w, r)
		return
		// }

		PageNotFound(w)
	})
}
