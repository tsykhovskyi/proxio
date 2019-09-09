package server

import (
	"bufio"
	"context"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

var Servers = make(map[string]*GroupServer)

type GroupServer struct {
	Server      *http.Server
	ForwardsMap map[string]*ForwardMap
}

type ForwardMap struct {
	Conn          *gossh.ServerConn
	ReqSshPayload remoteForwardRequest
}

func (gs *GroupServer) ServeHttp(w http.ResponseWriter, r *http.Request) {
	destAddr, destPortStr, _ := net.SplitHostPort(r.Host)
	if _, ok := gs.ForwardsMap[destAddr]; !ok {
		return
	}

	destPort, _ := strconv.Atoi(destPortStr)

	originAddr, orignPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(orignPortStr)

	payload := gossh.Marshal(&remoteForwardChannelData{
		DestAddr:   destAddr,
		DestPort:   uint32(destPort),
		OriginAddr: originAddr,
		OriginPort: uint32(originPort),
	})
	ch, reqs, err := gs.ForwardsMap[destAddr].Conn.OpenChannel(forwardedTCPChannelType, payload)
	if err != nil {
		log.Println(err)
		return
	}
	go gossh.DiscardRequests(reqs)

	err = r.WriteProxy(ch)
	if nil != err {
		panic(err)
	}

	bufRead := bufio.NewReader(ch)
	res, err := http.ReadResponse(bufRead, nil)
	if nil != err {
		panic(err)
	}

	for key, header := range res.Header {
		w.Header().Set(key, header[0])
	}
	io.Copy(w, res.Body)
}

func (gs *GroupServer) addChannel(host string, channel *gossh.ServerConn, reqPayload remoteForwardRequest) {
	gs.ForwardsMap[host] = &ForwardMap{
		Conn:          channel,
		ReqSshPayload: reqPayload,
	}
}

func (b *Balancer) AdjustNewForward(ctx context.Context, addr string, conn *gossh.ServerConn, reqPayload remoteForwardRequest) {
	host, port, _ := net.SplitHostPort(addr)
	if _, ok := Servers[port]; !ok {
		gs := &GroupServer{ForwardsMap: make(map[string]*ForwardMap, 0)}

		gs.Server = &http.Server{
			Addr:    ":" + port,
			Handler: &pHandler{httpForward: gs.ServeHttp},
		}
		// gs.Server.Shutdown(ctx)
		go func() {
			err := gs.Server.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()

		Servers[port] = gs
	}

	Servers[port].addChannel(host, conn, reqPayload)
}

type pHandler struct {
	httpForward func(http.ResponseWriter, *http.Request)
}

func (p *pHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.httpForward(w, r)
}
