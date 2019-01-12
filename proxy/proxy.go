package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

var (
	messageCounter int = 0
)

func NewProxy(localPort int, target string, storage *Storage) *Proxy {
	dest, err := url.Parse(target)
	if err != nil {
		panic("invalid target url")
	}

	return &Proxy{
		localPort,
		dest,
		make(chan *Message, 10),
		storage,
	}
}

type Proxy struct {
	localPort int
	dest      *url.URL
	Messages  chan *Message
	Storage   *Storage
}

func (p *Proxy) Serve() {
	hostListener, err := net.Listen("tcp", ":"+strconv.Itoa(p.localPort))
	if err != nil {
		panic(fmt.Sprintf("proxy lister error: %s", err))
	}

	proxy := httputil.NewSingleHostReverseProxy(p.dest)
	proxy.Transport = &transport{p, http.DefaultTransport}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	srv := &http.Server{}
	err = srv.Serve(hostListener)
	if err != nil {
		panic(fmt.Sprintf("unable to serve proxy: %s\n", err))
	}
}

type transport struct {
	*Proxy
	http.RoundTripper
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	message := newMessage(req)

	t.Proxy.handleMessageUpdate(message)

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		fmt.Println("Request cancelled", err.Error())
		return resp, err
	}

	message.StopTimer()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println("Error reading stream", err.Error())
		return resp, err
	}

	message.ResponseBody = bodyBytes
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	message.Response = resp
	t.Proxy.handleMessageUpdate(message)

	return resp, err
}

func (p *Proxy) handleMessageUpdate(message *Message) {
	if len(p.Messages) == cap(p.Messages) && len(p.Messages) > 0 {
		<-p.Messages
	}
	p.Messages <- message
	p.Storage.Add(message)
}
