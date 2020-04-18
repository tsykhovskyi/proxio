package server

import (
	"net/http"
	"strings"
)

func NewHttpServer(clientTrafficHandler, monitoringHandler http.Handler) *http.Server {
	splitHandler := SubDomainMiddleware(
		clientTrafficHandler,
		// http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 	w.WriteHeader(http.StatusNotFound)
		// 	w.Write([]byte("<h1>Forward path not found</h1>"))
		// }),
		monitoringHandler,
	)

	return &http.Server{
		Addr:    ":80",
		Handler: splitHandler,
	}
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
