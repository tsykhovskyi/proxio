package server

import "net/http"

const proxyNotFoundHtml = `
<h1>Proxy not found</h1>
We were unable to found proxy for this domain
`

const remoteServerNotFoundHtml = `
<h1>Destination server was not found</h1>
Maybe you forgot to launch your local server to which you suppose to proxy traffic
`

func ProxyNotFound(w http.ResponseWriter) {
	httpError(w, 404, proxyNotFoundHtml)
}

func RemoteServerNotFound(w http.ResponseWriter) {
	httpError(w, 500, remoteServerNotFoundHtml)
}

func httpError(w http.ResponseWriter, statusCode int, error string) {
	_, err := w.Write([]byte(error))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err != nil {
		panic(err)
	}
}
