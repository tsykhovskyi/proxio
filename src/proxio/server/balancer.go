package server

import (
	"strings"
)

type ForwardDomain string

func (domain ForwardDomain) Subdomain(originDomain string) string {
	return strings.TrimSuffix(string(domain), originDomain)
}

type Balancer struct {
	Host        string
	forwardsMap map[ForwardDomain]*Proxy
}

type Proxy struct {
	RequestedAddr string
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

func (b *Balancer) CreateNewForward(addr string, port uint32, tunnel *SshTunnel) ForwardDomain {
	desiredDomain := ForwardDomain(addr)
	subdomain := desiredDomain.Subdomain(b.Host)
	if subdomain == "" {
		subdomain = tunnel.user
	}

	realaddr := ForwardDomain(subdomain + "." + b.Host)
	if _, ok := b.forwardsMap[realaddr]; ok {
		panic("error while generating domain")
	}

	b.forwardsMap[(realaddr)] = &Proxy{
		RequestedAddr: addr,
		Port:          port,
		TunnelId:      tunnel.sessionId,
	}

	return realaddr
}

func (b *Balancer) DeleteForwardForSession(sessionId string) {
	for key, proxy := range b.forwardsMap {
		if proxy.TunnelId == sessionId {
			delete(b.forwardsMap, key)
			return
		}
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
