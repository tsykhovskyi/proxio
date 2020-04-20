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

func (b *Balancer) AdjustNewForward(ctx context.Context, addr string, port uint32, tunnelId string) {
	b.forwardsMap[ForwardAddress(addr)] = &ForwardDest{
		Addr:     addr,
		Port:     port,
		TunnelId: tunnelId,
	}
}

func (b *Balancer) GetByAddress(addr string) *ForwardDest {
	if dest := b.forwardsMap[ForwardAddress(addr)]; dest != nil {
		return dest
	}
	return nil
}

func NewBalancer() *Balancer {
	return &Balancer{
		forwardsMap: make(map[ForwardAddress]*ForwardDest, 0),
	}
}
