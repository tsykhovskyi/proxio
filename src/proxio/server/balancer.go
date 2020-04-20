package server

import (
	"context"
	"net/http"
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
	Addr     string
	Port     uint32
	TunnelId string
}

// type Tunnel interface {
// 	// Id() string
// 	// GetChannel(DestAddr string, DestPort uint32, OriginAddr string, OriginPort uint32) io.GetChannel
// }

func (b *Balancer) HasTunnelOnPort(port uint32, tunnelId string) bool {
	return b.getByTunnel(tunnelId) != nil
}

func (b *Balancer) AdjustNewForward(ctx context.Context, addr string, port uint32, tunnelId string) {
	b.addConn(addr, port, tunnelId)
}

func (b *Balancer) addConn(addr string, port uint32, tunnelId string) {
	b.forwardsMap[ForwardAddress(addr)] = &ForwardDest{
		Addr:     addr,
		Port:     port,
		TunnelId: tunnelId,
	}
}

func (b *Balancer) getByTunnel(tunnel string) *ForwardDest {
	for _, fw := range b.forwardsMap {
		if fw.TunnelId == tunnel {
			return fw
		}
	}
	return nil
}

func (b *Balancer) GetByAddress(addr string) *ForwardDest {
	if dest := b.forwardsMap[ForwardAddress(addr)]; dest != nil {
		return dest
	}
	return nil
}

func NewBalancer() *Balancer {
	balancer := &Balancer{
		forwardsMap: make(map[ForwardAddress]*ForwardDest, 0),
	}
	// balancer.httpHandler = http.HandlerFunc(balancer.ServeHTTP)

	return balancer
}
