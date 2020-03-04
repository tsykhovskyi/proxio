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

type ForwardServers struct {
	servers map[uint32]*ForwardServer
}

type ForwardAddress string

type ForwardServer struct {
	Server      *http.Server
	ForwardsMap map[ForwardAddress]*ForwardDest
}

type ForwardDest struct {
	Conn          *gossh.ServerConn
	ReqSshPayload remoteForwardRequest
}

func (fss *ForwardServers) HasConnectionOnPort(conn *gossh.ServerConn, port uint32) bool {
	return fss.servers[port] != nil && fss.servers[port].hasChannel(conn)
}

func (fss *ForwardServers) HasForwardAddress(address string, port uint32) bool {
	return fss.servers[port] != nil && fss.servers[port].ForwardsMap[ForwardAddress(address)] != nil
}

func (fss *ForwardServers) AdjustNewForward(ctx context.Context, addr string, port uint32, conn *gossh.ServerConn, reqPayload remoteForwardRequest) {
	if _, ok := fss.servers[port]; !ok {
		fs := &ForwardServer{ForwardsMap: make(map[ForwardAddress]*ForwardDest, 0)}

		fs.Server = &http.Server{
			Addr:    ":" + strconv.Itoa(int(port)),
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fs.ServeHttp(w, r) }),
		}
		// _ = fs.Server.Shutdown(ctx)
		go func() {
			err := fs.Server.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}()

		fss.servers[port] = fs
	}

	fss.servers[port].addChannel(addr, conn, reqPayload)
}

func (fs *ForwardServer) ServeHttp(w http.ResponseWriter, r *http.Request) {
	destAddr, destPortStr, _ := net.SplitHostPort(r.Host)
	if _, ok := fs.ForwardsMap[ForwardAddress(destAddr)]; !ok {
		return
	}

	destPort, _ := strconv.Atoi(destPortStr)

	originAddr, orignPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(orignPortStr)

	payload := gossh.Marshal(&remoteForwardChannelData{
		// DestAddr:   destAddr,
		DestAddr:   "127.0.0.1",
		DestPort:   uint32(destPort),
		OriginAddr: originAddr,
		OriginPort: uint32(originPort),
	})
	ch, reqs, err := fs.ForwardsMap[ForwardAddress(destAddr)].Conn.OpenChannel(forwardedTCPChannelType, payload)
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
	_, err = io.Copy(w, res.Body)
	if err != nil {
		panic(err)
	}
}

func (fs *ForwardServer) addChannel(host string, channel *gossh.ServerConn, reqPayload remoteForwardRequest) {
	fs.ForwardsMap[ForwardAddress(host)] = &ForwardDest{
		Conn:          channel,
		ReqSshPayload: reqPayload,
	}
}

func (fs *ForwardServer) hasChannel(conn *gossh.ServerConn) bool {
	for _, fw := range fs.ForwardsMap {
		if fw.Conn == conn {
			return true
		}
	}
	return false
}

func NewForwardServers() *ForwardServers {
	return &ForwardServers{servers: make(map[uint32]*ForwardServer)}
}
