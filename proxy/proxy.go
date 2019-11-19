package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var (
	messageCounter int = 0
)

func ListenAndServe(listener net.Listener, target string) chan *Message {
	targetUrl, err := url.Parse(target)
	if err != nil {
		panic("invalid target url")
	}

	proxy := &Proxy{
		listener,
		targetUrl,
		make(chan *Message, 1),
	}
	return proxy.Serve()
}

type Proxy struct {
	listener net.Listener
	target   *url.URL
	messages chan *Message
}

func (p *Proxy) Serve() chan *Message {
	go func() {
		proxy := httputil.NewSingleHostReverseProxy(p.target)
		proxy.Transport = &transport{p, http.DefaultTransport}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			proxy.ServeHTTP(w, r)
		})

		srv := &http.Server{}
		err := srv.Serve(p.listener)
		if err != nil {
			panic(fmt.Sprintf("unable to serve proxy: %s\n", err))
		}
	}()

	return p.messages
}

type transport struct {
	*Proxy
	http.RoundTripper
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	message := newMessage(req)

	if nil != req.Body {
		bodyBytes, err := ioutil.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			fmt.Println("Error reading request stream", err.Error())
			return resp, err
		}
		message.RequestBody = bodyBytes
		req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	t.Proxy.handleMessageUpdate(message)

	resp, err = t.RoundTripper.RoundTrip(req)
	if err != nil {
		message.Cancel()
		t.Proxy.handleMessageUpdate(message)
		return resp, err
	}

	message.StopTimer()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		fmt.Println("Error reading response stream", err.Error())
		return resp, err
	}
	message.ResponseBody = bodyBytes
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	message.Response = resp
	t.Proxy.handleMessageUpdate(message)

	return resp, err
}

func (p *Proxy) handleMessageUpdate(message *Message) {
	if len(p.messages) == cap(p.messages) && len(p.messages) > 0 {
		<-p.messages
	}
	p.messages <- message
}
