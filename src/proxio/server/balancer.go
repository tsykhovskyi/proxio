package server

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
)

type Balancer struct {
	httpHandler http.Handler
	forwardsMap map[ForwardAddress]*ForwardDest
}

type ForwardAddress string

type HttpForwardServer struct {
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

func (b *Balancer) HasTunnelOnPort(port uint32, tunnel Tunnel) bool {
	return b.getByTunnel(tunnel) != nil
}

func (b *Balancer) UpdatePayloadConnectionOnPort(tunnel Tunnel, addr string, port uint32) {
	dest := b.getByTunnel(tunnel)
	dest.Addr = addr
	dest.Port = port
}

func (b *Balancer) AdjustNewForward(ctx context.Context, addr string, port uint32, tunnel Tunnel) {
	b.addConn(addr, port, tunnel)
}

func (b *Balancer) ServeHttp(w http.ResponseWriter, r *http.Request) {
	destAddr := r.Host

	if _, ok := b.forwardsMap[ForwardAddress(destAddr)]; !ok {
		return
	}

	originAddr, originPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(originPortStr)

	forward := b.forwardsMap[ForwardAddress(destAddr)]

	tunnel := forward.Tunnel

	ch := tunnel.ReadWriteCloser(forward.Addr, forward.Port, originAddr, uint32(originPort))

	err := r.WriteProxy(ch)
	if nil != err {
		panic(err)
	}

	bufRead := bufio.NewReader(ch)
	res, err := http.ReadResponse(bufRead, r)
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

func (b *Balancer) addConn(addr string, port uint32, tunnel Tunnel) {
	b.forwardsMap[ForwardAddress(addr)] = &ForwardDest{
		Addr:   addr,
		Port:   port,
		Tunnel: tunnel,
	}
}

func (b *Balancer) getByTunnel(tunnel Tunnel) *ForwardDest {
	for _, fw := range b.forwardsMap {
		if fw.Tunnel.Id() == tunnel.Id() {
			return fw
		}
	}
	return nil
}

func NewBalancer() *Balancer {
	balancer := &Balancer{
		forwardsMap: make(map[ForwardAddress]*ForwardDest, 0),
	}
	balancer.httpHandler = http.HandlerFunc(balancer.ServeHttp)

	return balancer
}
