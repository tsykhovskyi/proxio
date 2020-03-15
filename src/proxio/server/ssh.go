package server

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"strconv"
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

type SSHForwardHandler struct {
	balancer   *Balancer
	port       uint32
	privateKey string
}

func (h *SSHForwardHandler) Start() error {
	s := &ssh.Server{
		Addr: ":" + strconv.Itoa(int(h.port)),
		Handler: func(session ssh.Session) {
			key := gossh.MarshalAuthorizedKey(session.PublicKey())
			out := fmt.Sprintf("Hi, %s\n", key)
			io.WriteString(session, out)
			select {}
		},
		LocalPortForwardingCallback: func(ctx ssh.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		SessionRequestCallback: func(sess ssh.Session, requestType string) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
			return true
		},
		PtyCallback: func(ctx ssh.Context, pty ssh.Pty) bool {
			return false
		},
		PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
			return true
		},
	}
	s.AddHostKey(publicKeyFile(h.privateKey))

	s.RequestHandlers = map[string]ssh.RequestHandler{
		"prepare-tcpip-forward": h.handleSSHRequest, // custom type, sending from client application
		"tcpip-forward":         h.handleSSHRequest,
		"cancel-tcpip-forward":  h.handleSSHRequest,
	}

	s.ChannelHandlers = ssh.DefaultChannelHandlers
	s.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	return s.ListenAndServe()
}

func (h *SSHForwardHandler) handleSSHRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	conn := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	tunnel := &SshTunnel{conn}

	switch req.Type {
	case "prepare-tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}

		h.balancer.AdjustNewForward(ctx, reqPayload.BindAddr, reqPayload.BindPort, tunnel)

		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}

		// problem with ssh lib that send 127.0.0.1 instead of localhost
		if h.balancer.HasTunnelOnPort(reqPayload.BindPort, tunnel) {
			h.balancer.UpdatePayloadConnectionOnPort(tunnel, reqPayload.BindAddr, reqPayload.BindPort)
			return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
		}

		h.balancer.AdjustNewForward(ctx, reqPayload.BindAddr, reqPayload.BindPort, tunnel)
		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})

	case "cancel-tcpip-forward":
		var reqPayload remoteForwardCancelRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		// addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		// ln, ok := h.forwards[addr]
		// h.Unlock()
		// if ok {
		// 	ln.Close()
		// }

		return true, nil
	default:
		return false, nil
	}
}

type SshTunnel struct {
	*gossh.ServerConn
}

func (st *SshTunnel) Id() string {
	return string(st.SessionID())
}

func (st *SshTunnel) ReadWriteCloser(destAddr string, destPort uint32, originAddr string, originPort uint32) io.ReadWriteCloser {
	payload := gossh.Marshal(&remoteForwardChannelData{
		DestAddr:   destAddr,
		DestPort:   destPort,
		OriginAddr: originAddr,
		OriginPort: originPort,
	})

	channel, reqs, err := st.OpenChannel(forwardedTCPChannelType, payload)
	if err != nil {
		panic(err)
	}
	go gossh.DiscardRequests(reqs)

	return channel
}

func NewSshForwardHandler(balancer *Balancer, port uint32, privateKey string) *SSHForwardHandler {
	b := &SSHForwardHandler{
		balancer:   balancer,
		port:       port,
		privateKey: privateKey,
	}

	return b
}
