package server

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

type Balancer struct {
	servers map[uint32]*ForwardServer
}

type ForwardAddress string

type ForwardServer struct {
	Server      *http.Server
	ForwardsMap map[ForwardAddress]*ForwardDest
}

type ForwardDest struct {
	Addr   string
	Port   uint32
	Tunnel Tunnel
}

type Tunnel interface {
	Id() string
	ReadWriteCloser(DestAddr string, DestPort uint32, OriginAddr string, OriginPort uint32) io.ReadWriteCloser
}

func (fss *Balancer) HasTunnelOnPort(port uint32, tunnel Tunnel) bool {
	return fss.servers[port] != nil && fss.servers[port].getByTunnel(tunnel) != nil
}

func (fss *Balancer) UpdatePayloadConnectionOnPort(port uint32, reqPayload remoteForwardRequest, tunnel Tunnel) {
	dest := fss.servers[port].getByTunnel(tunnel)
	dest.Addr = reqPayload.BindAddr
	dest.Port = reqPayload.BindPort
}

func (fss *Balancer) AdjustNewForward(ctx context.Context, addr string, port uint32, tunnel Tunnel) {
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

	fss.servers[port].addConn(addr, port, tunnel)
}

func (fs *ForwardServer) ServeHttp(w http.ResponseWriter, r *http.Request) {
	destAddr, _, _ := net.SplitHostPort(r.Host)

	if _, ok := fs.ForwardsMap[ForwardAddress(destAddr)]; !ok {
		return
	}

	originAddr, originPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(originPortStr)

	forward := fs.ForwardsMap[ForwardAddress(destAddr)]

	tunnel := forward.Tunnel

	ch := tunnel.ReadWriteCloser(forward.Addr, forward.Port, originAddr, uint32(originPort))

	err := r.WriteProxy(ch)
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

func (fs *ForwardServer) addConn(addr string, port uint32, tunnel Tunnel) {
	fs.ForwardsMap[ForwardAddress(addr)] = &ForwardDest{
		Addr:   addr,
		Port:   port,
		Tunnel: tunnel,
	}
}

func (fs *ForwardServer) getByTunnel(tunnel Tunnel) *ForwardDest {
	for _, fw := range fs.ForwardsMap {
		if fw.Tunnel.Id() == tunnel.Id() {
			return fw
		}
	}
	return nil
}

func NewBalancer() *Balancer {
	return &Balancer{servers: make(map[uint32]*ForwardServer)}
}
