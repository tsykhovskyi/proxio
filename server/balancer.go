package server

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"net"
	"strconv"
	"sync"
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
	forwards map[string]net.Listener

	sync.Mutex
}

func (b *Balancer) HandleSSHRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	b.Lock()
	if b.forwards == nil {
		b.forwards = make(map[string]net.Listener)
	}
	b.Unlock()
	conn := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	switch req.Type {
	case "prepare-tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		b.AdjustNewForward(ctx, addr, conn, reqPayload)

		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		if gs := Servers[fmt.Sprintf("%d", reqPayload.BindPort)]; gs != nil && gs.hasChannel(conn) {
			return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
		}

		addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		fmt.Println(addr)

		b.AdjustNewForward(ctx, addr, conn, reqPayload)
		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})

	case "cancel-tcpip-forward":
		var reqPayload remoteForwardCancelRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		b.Lock()
		ln, ok := b.forwards[addr]
		b.Unlock()
		if ok {
			ln.Close()
		}
		return true, nil
	default:
		return false, nil
	}
}
