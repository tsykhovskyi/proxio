package server

import "net/http"

const pageNotFound = `
<h1>Page not found</h1>
Go to home page to continue
`

const proxyNotFoundHtml = `
<h1>Proxy not found</h1>
We were unable to found proxy for this domain
`

const remoteServerNotFoundHtml = `
<h1>Destination server was not found</h1>
Maybe you forgot to launch your local server to which you suppose to proxy traffic
`

func PageNotFound(w http.ResponseWriter) {
	httpError(w, 404, pageNotFound)
}

func ProxyNotFound(w http.ResponseWriter) {
	httpError(w, 404, proxyNotFoundHtml)
}

func RemoteServerNotFound(w http.ResponseWriter) {
	httpError(w, 500, remoteServerNotFoundHtml)
}

func httpError(w http.ResponseWriter, statusCode int, error string) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err := w.Write([]byte(error))
	if err != nil {
		panic(err)
	}
}
