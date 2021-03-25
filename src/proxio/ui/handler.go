package ui

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"proxio/client"
	"proxio/repository"
	"proxio/ssh"
)

type contextKey struct {
	name string
}

var (
	// Value of type Session
	ContextKeyRequestSession = &contextKey{name: "session"}
)

// This will check authentication header and attach session object if user authenticated
func NewSessionMiddleware(sessions repository.Sessions) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer next.ServeHTTP(w, r)

			sessionId := r.Header.Get("Authentication")
			if sessionId == "" {
				return
			}
			if session, ok := sessions.Find(sessionId); ok {
				context.Set(r, ContextKeyRequestSession, session)
			}
		})
	}
}

// This will compare user token to attempted domain action
// and return 401 if not allowed
func NewDomainPermissionMiddleware(balancer *ssh.Balancer) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			domain := r.URL.Query().Get("domain")
			if domain == "" {
				next.ServeHTTP(w, r)
				return
			}

			accessGranted := false

			token := r.URL.Query().Get("token")
			if token != "" {
				accessGranted = balancer.TestDomainToken(domain, token)
			}

			if !accessGranted {
				http.Error(w, "Incorrect token or domain is invalid", 403)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Traffic request handler
func NewTrafficRequestHandler(storage *Storage) *TrafficRequestHandler {
	return &TrafficRequestHandler{
		Storage: storage,
	}
}

type TrafficRequestHandler struct {
	Storage *Storage
}

func (c *TrafficRequestHandler) domainTraffic(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "Domain not provided", 400)
	}
	messages := c.Storage.All(domain)

	response := make([]*client.MessageContent, len(messages))

	for i, m := range messages {
		response[i] = m.GetContext()
	}

	payload, err := json.Marshal(response)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on message reading: %s", err), 500)
	}

	w.Write(payload)
}

func (c *TrafficRequestHandler) clear(w http.ResponseWriter, r *http.Request) {
	c.Storage.RemoveAll("")
}

func (c *TrafficRequestHandler) session(w http.ResponseWriter, r *http.Request) {
	val := context.Get(r, ContextKeyRequestSession)
	if val == nil {
		http.Error(w, "Unauthorized", 401)
		return
	}
	session := val.(repository.Session)

	payload, err := json.Marshal(session)
	if err != nil {
		http.Error(w, "Something goes wrong", 500)
	}
	w.Write(payload)
}

// Websocket handler
func NewWsHandler(connectionPool *Pool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connection := serveWs(w, r, connectionPool.closeChan)

		domain := r.URL.Query().Get("domain")
		if domain == "" {
			http.Error(w, "Domain not provided", 400)
		}
		connectionPool.NewConnection(domain, connection)
	})
}

// SPA handler
func NewSpaHandler() http.Handler {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return spaHandler{staticPath: wd + "/../telemetry/dist/telemetry", indexPath: "index.html"}
}

type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	path = filepath.Join(h.staticPath, path)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		indexFullPath := filepath.Join(h.staticPath, h.indexPath)
		http.ServeFile(w, r, indexFullPath)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// Proxy handler
// for debug purpose
func NewProxyHandler(addr string) http.Handler {
	proxyPath, _ := url.Parse(addr)
	return httputil.NewSingleHostReverseProxy(proxyPath)
}
