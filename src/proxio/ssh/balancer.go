package ssh

import (
	"errors"
	"fmt"
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

func (b *Balancer) ValidateRequestDomain(addr string, port uint32) (bool, error) {
	if port != 80 {
		return false, errors.New("Only 80 port can be forwarded")
	}
	if addr == "" || addr == b.Host || addr == "localhost" {
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
	if addr == "" || addr == b.Host || addr == "localhost" {
		return ""
	}
	return strings.TrimSuffix(addr, "."+b.Host)
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
		Domain:        domain,
		Proto:         "http",
		RequestedAddr: addr,
		Port:          port,
		Tunnel:        tunnel,
	}

	return domain, nil
}

func (b *Balancer) GetProxyBySessionId(sessionId string) *Proxy {
	for _, proxy := range b.forwardsMap {
		if proxy.Tunnel.sessionId == sessionId {
			return proxy
		}
	}
	return nil
}

func (b *Balancer) DeleteProxyForSession(sessionId string) {
	for key, proxy := range b.forwardsMap {
		if proxy.Tunnel.sessionId == sessionId {
			delete(b.forwardsMap, key)
			return
		}
	}
}

func (b *Balancer) GetProxyByAddress(addr string) *Proxy {
	if dest := b.forwardsMap[Domain(addr)]; dest != nil {
		return dest
	}
	return nil
}

func (b *Balancer) TestDomainToken(domain string, token string) bool {
	proxy := b.GetProxyByAddress(domain)
	if proxy == nil {
		return false
	}
	return proxy.Tunnel.sessionId == token
}

func (b *Balancer) TestDomainPublicKey(domain string, publicKey string) bool {
	proxy := b.GetProxyByAddress(domain)
	if proxy == nil {
		return false
	}
	return proxy.Tunnel.publicKeyStr == publicKey
}

type Proxy struct {
	Domain        Domain
	Proto         string
	RequestedAddr string
	Port          uint32
	Tunnel        *SshTunnel
}

func (p *Proxy) Host() string {
	return fmt.Sprintf("%s://%s", p.Proto, p.Domain)
}

func NewBalancer(host string, suggester Suggester) *Balancer {
	return &Balancer{
		Host:        host,
		suggester:   suggester,
		forwardsMap: make(map[Domain]*Proxy, 0),
	}
}
