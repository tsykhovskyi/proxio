package server

import (
	"bufio"
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"net"
	"net/http"
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

type SSHForwardServer struct {
	balancer   *Balancer
	port       uint32
	privateKey string
	tunnels    map[string]*SshTunnel
}

func (sfs *SSHForwardServer) Start() error {
	s := &ssh.Server{
		Addr: ":" + strconv.Itoa(int(sfs.port)),
		Handler: func(session ssh.Session) {
			key := gossh.MarshalAuthorizedKey(session.PublicKey())
			out := fmt.Sprintf("Hi, %s\n", key)
			// session.Write([]byte(out))
			conn := session.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)
			fmt.Printf("%s\n", string(conn.SessionID()))
			io.WriteString(session, out)
			// todo handle saving session related to conn
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
	s.AddHostKey(publicKeyFile(sfs.privateKey))

	s.RequestHandlers = map[string]ssh.RequestHandler{
		"prepare-tcpip-forward": sfs.handleSSHRequest, // custom type, sending from client application
		"tcpip-forward":         sfs.handleSSHRequest,
		"cancel-tcpip-forward":  sfs.handleSSHRequest,
	}

	s.ChannelHandlers = ssh.DefaultChannelHandlers
	s.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	return s.ListenAndServe()
}

func (sfs *SSHForwardServer) handleSSHRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	conn := ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn)
	// sessionId := ctx.Value(ssh.ContextKeySessionID).(string)
	tunnel := &SshTunnel{conn: conn}

	switch req.Type {
	// case "prepare-tcpip-forward":
	// 	var reqPayload remoteForwardRequest
	// 	if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
	// 		return false, []byte{}
	// 	}
	//
	// 	sfs.balancer.AdjustNewForward(ctx, reqPayload.BindAddr, reqPayload.BindPort, tunnel)
	//
	// 	return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}

		// problem with ssh lib that send 127.0.0.1 instead of localhost
		// if sfs.balancer.HasTunnelOnPort(reqPayload.BindPort, tunnel) {
		// 	sfs.balancer.UpdatePayloadConnectionOnPort(tunnel, reqPayload.BindAddr, reqPayload.BindPort)
		// 	return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
		// }

		sfs.addTunnel(tunnel)
		sfs.balancer.AdjustNewForward(ctx, reqPayload.BindAddr, reqPayload.BindPort, tunnel.Id())
		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})

	case "cancel-tcpip-forward":
		var reqPayload remoteForwardCancelRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}
		// addr := net.JoinHostPort(reqPayload.BindAddr, strconv.Itoa(int(reqPayload.BindPort)))
		// ln, ok := sfs.forwards[addr]
		// sfs.Unlock()
		// if ok {
		// 	ln.Close()
		// }

		return true, nil
	default:
		return false, nil
	}
}

func (sfs *SSHForwardServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	destAddr := r.Host

	dest := sfs.balancer.GetByAddress(destAddr)
	if nil == dest {
		return
	}

	originAddr, originPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(originPortStr)

	tunnel := sfs.getTunnel(dest.TunnelId)

	ch := tunnel.GetChannel(dest.Addr, dest.Port, originAddr, uint32(originPort))

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

func (sfs *SSHForwardServer) addTunnel(tunnel *SshTunnel) {
	sfs.tunnels[tunnel.Id()] = tunnel
}

func (sfs *SSHForwardServer) getTunnel(id string) *SshTunnel {
	return sfs.tunnels[id]
}

type SshTunnel struct {
	conn    *gossh.ServerConn
	session ssh.Session
}

func (st *SshTunnel) Id() string {
	return string(st.conn.SessionID())
}

func (st *SshTunnel) GetChannel(destAddr string, destPort uint32, originAddr string, originPort uint32) io.ReadWriteCloser {
	payload := gossh.Marshal(&remoteForwardChannelData{
		DestAddr:   destAddr,
		DestPort:   destPort,
		OriginAddr: originAddr,
		OriginPort: originPort,
	})

	channel, reqs, err := st.conn.OpenChannel(forwardedTCPChannelType, payload)
	sessId := st.conn.SessionID()
	fmt.Printf("%s\n", sessId)

	if err != nil {
		panic(err)
	}
	go gossh.DiscardRequests(reqs)

	return channel
}

func NewSshForwardServer(balancer *Balancer, port uint32, privateKey string) *SSHForwardServer {
	b := &SSHForwardServer{
		balancer:   balancer,
		port:       port,
		privateKey: privateKey,
		tunnels:    make(map[string]*SshTunnel),
	}

	return b
}
