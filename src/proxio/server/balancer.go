package server

import (
	"errors"
	"regexp"
	"strings"
)

type Domain string

func (domain Domain) Subdomain(originDomain string) string {
	return strings.TrimSuffix(string(domain), originDomain)
}

type Balancer struct {
	Host        string
	suggester   Suggester
	forwardsMap map[Domain]*Proxy
}

type Proxy struct {
	RequestedAddr string
	Port          uint32
	TunnelId      string
}

func (b *Balancer) ValidateRequestDomain(addr string, port uint32) (bool, error) {
	if port != 80 {
		return false, errors.New("Only 80 port can be forwarded")
	}
	if addr == "" || addr == b.Host {
		return true, nil
	}

	if !strings.HasSuffix(addr, "."+b.Host) {
		return false, errors.New("Host should be subdomain of " + b.Host)
	}

	subdomain := strings.TrimSuffix(addr, "."+b.Host)
	if match, _ := regexp.Match(`^[a-z0-9]{0,60}$`, []byte(subdomain)); match == false {
		return false, errors.New("Subdomain should contain only a-z letters and/or 0-9 digits")
	}

	return true, nil
}

func (b *Balancer) Subdomain(addr string) string {
	addr = strings.TrimSuffix(addr, b.Host)
	if addr == "" {
		return ""
	}
	return strings.TrimSuffix(addr, ".")
}

func (b *Balancer) CreateNewForward(addr string, port uint32, tunnel *SshTunnel) (Domain, error) {
	if _, err := b.ValidateRequestDomain(addr, port); err != nil {
		return "", err
	}

	desiredDomain := b.Subdomain(addr)

	if desiredDomain == "" {
		domainSuggested := false
		for i := 0; i < 3; i++ {
			desiredDomain = b.suggester.Suggest(tunnel, i)
			if _, ok := b.forwardsMap[Domain(desiredDomain+"."+b.Host)]; !ok {
				domainSuggested = true
				break
			}
		}
		if !domainSuggested {
			return "", errors.New("Unable to provide domain")
		}
	}

	domain := Domain(desiredDomain + "." + b.Host)
	if _, ok := b.forwardsMap[domain]; ok {
		return "", errors.New("Unable to provide desired domain: " + string(domain))
	}

	b.forwardsMap[domain] = &Proxy{
		RequestedAddr: addr,
		Port:          port,
		TunnelId:      tunnel.sessionId,
	}

	return domain, nil
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
	if dest := b.forwardsMap[Domain(addr)]; dest != nil {
		return dest
	}
	return nil
}

func NewBalancer(host string, suggester Suggester) *Balancer {
	return &Balancer{
		Host:        host,
		suggester:   suggester,
		forwardsMap: make(map[Domain]*Proxy, 0),
	}
}
