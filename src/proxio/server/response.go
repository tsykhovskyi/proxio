package server

import "net/http"

const PROXY_NOT_FOUND = `
<h1>Proxy not found</h1>
We were unable to found proxy for this domain
`

const REMOTE_SERVER_NOT_FOUND = `
<h1>Destination server was not found</h1>
Maybe you forgot to launch your local server to which you suppose to proxy traffic
`

func ProxyNotFound(w http.ResponseWriter) {
	http.Error(w, "Proxy not found", 404)
	// _, err := w.Write([]byte(PROXY_NOT_FOUND))
	// if err != nil {
	// 	panic(err)
	// }
}
