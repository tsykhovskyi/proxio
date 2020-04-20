package server

import (
	"bufio"
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"net"
	"net/http"
	"proxio/client"
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
	tracker    *client.TrafficTracker
	port       uint32
	privateKey string
	tunnels    map[string]*SshTunnel
}

func (sfs *SSHForwardServer) Start() error {
	s := &ssh.Server{
		Addr: ":" + strconv.Itoa(int(sfs.port)),
		Handler: func(session ssh.Session) {
			_ = gossh.MarshalAuthorizedKey(session.PublicKey())

			conn := session.Context().Value(ssh.ContextKeyConn).(*gossh.ServerConn)
			tunnel := sfs.getTunnel(string(conn.SessionID()))
			if tunnel == nil {
				panic("tunnel not found")
			}
			tunnel.session = session

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
	tunnel := &SshTunnel{conn: conn}

	switch req.Type {
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			// TODO: log parse failure
			return false, []byte{}
		}

		sfs.addTunnel(tunnel)
		sfs.balancer.AdjustNewForward(reqPayload.BindAddr, reqPayload.BindPort, tunnel.Id())
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
	dest := sfs.balancer.GetByAddress(r.Host)
	if nil == dest {
		ProxyNotFound(w)
		return
	}
	tunnel := sfs.getTunnel(dest.TunnelId)
	if nil == tunnel {
		panic("tunnel not found while ssh-forwarding")
	}

	sfs.tracker.RequestStarted(r)
	tunnel.session.Write([]byte("You have serve traffic\n"))

	originAddr, originPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(originPortStr)

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

	sfs.tracker.RequestFinished(r, res)

	for key, header := range res.Header {
		w.Header().Set(key, header[0])
	}
	w.WriteHeader(res.StatusCode)
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

func NewSshForwardServer(balancer *Balancer, tracker *client.TrafficTracker, port uint32, privateKey string) *SSHForwardServer {
	b := &SSHForwardServer{
		balancer:   balancer,
		port:       port,
		privateKey: privateKey,
		tracker:    tracker,
		tunnels:    make(map[string]*SshTunnel),
	}

	return b
}
