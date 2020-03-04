package server

import (
	"github.com/gliderlabs/ssh"
)
import gossh "golang.org/x/crypto/ssh"

const (
	forwardedTCPChannelType = "forwarded-tcpip"
)

type remoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

type remoteForwardChannelData struct {
	DestAddr   string
	DestPort   uint32
	OriginAddr string
	OriginPort uint32
}

type remoteForwardSuccess struct {
	BindPort uint32
}

type remoteForwardCancelRequest struct {
	BindAddr string
	BindPort uint32
}

type Balancer struct {
	servers *ForwardServers
}

func (b *Balancer) HandleSSHRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	conn := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	switch req.Type {
	case "prepare-tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}

		b.servers.AdjustNewForward(ctx, reqPayload.BindAddr, reqPayload.BindPort, conn, reqPayload)

		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}

		// problem with ssh lib that send 127.0.0.1 instead of localhost
		if b.servers.HasConnectionOnPort(conn, reqPayload.BindPort) {
			return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
		}

		b.servers.AdjustNewForward(ctx, reqPayload.BindAddr, reqPayload.BindPort, conn, reqPayload)
		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})

	case "cancel-tcpip-forward":
		var reqPayload remoteForwardCancelRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		// addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		// ln, ok := b.forwards[addr]
		// b.Unlock()
		// if ok {
		// 	ln.Close()
		// }
		return true, nil
	default:
		return false, nil
	}
}

func NewBalancer(servers *ForwardServers) *Balancer {
	b := &Balancer{
		servers: servers,
	}

	return b
}
