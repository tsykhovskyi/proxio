package server

import (
	"strings"
)

type ForwardDomain string

func (domain ForwardDomain) Subdomain(originDomain string) string {
	return ""
}

type Balancer struct {
	Host        string
	forwardsMap map[ForwardDomain]*Proxy
}

type Proxy struct {
	RequestedAddr string
	RealAddr      string
	Port          uint32
	TunnelId      string
}

func (b *Balancer) ValidateRequestDomain(addr string, port uint32) (bool, string) {
	if port != 80 {
		return false, "Only 80 port can be forwarded"
	}
	if !strings.HasSuffix(addr, b.Host) {
		return false, "Host should be subdomain of " + b.Host
	}

	return true, ""
}

func (b *Balancer) AdjustNewForward(addr string, port uint32, tunnelId string) {
	realaddr := "subdomain2.localhost"
	b.forwardsMap[ForwardDomain(realaddr)] = &Proxy{
		RequestedAddr: addr,
		RealAddr:      realaddr,
		Port:          port,
		TunnelId:      tunnelId,
	}
}

func (b *Balancer) GetByAddress(addr string) *Proxy {
	if dest := b.forwardsMap[ForwardDomain(addr)]; dest != nil {
		return dest
	}
	return nil
}

func NewBalancer(host string) *Balancer {
	return &Balancer{
		Host:        host,
		forwardsMap: make(map[ForwardDomain]*Proxy, 0),
	}
}
