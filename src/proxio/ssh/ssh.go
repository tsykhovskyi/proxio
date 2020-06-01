package ssh

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
	port         uint32
	privateKey   string
	uiDomain     string
	balancer     *Balancer
	tracker      *client.TrafficTracker
	tunnels      map[string]*SshTunnel
	tunnelErrors map[string]error
}

func (sfs *SSHForwardServer) Start() error {
	s := &ssh.Server{
		Addr:    ":" + strconv.Itoa(int(sfs.port)),
		Handler: sfs.HandleSshSession,
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
		"tcpip-forward":        sfs.HandleSSHRequest,
		"cancel-tcpip-forward": sfs.HandleSSHRequest,
	}

	s.ChannelHandlers = ssh.DefaultChannelHandlers
	s.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	return s.ListenAndServe()
}

func (sfs *SSHForwardServer) HandleSSHRequest(ctx ssh.Context, srv *ssh.Server, req *gossh.Request) (bool, []byte) {
	switch req.Type {
	case "tcpip-forward":
		var reqPayload remoteForwardRequest
		if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
			return false, []byte{}
		}
		err := sfs.addTunnel(ctx, reqPayload)
		if err != nil {
			return false, nil
		}

		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})
	default:
		return false, nil
	}
}

func (sfs *SSHForwardServer) HandleSshSession(s ssh.Session) {
	session := &Session{s}

	_ = gossh.MarshalAuthorizedKey(session.PublicKey())
	sessionId := session.Context().Value(ssh.ContextKeySessionID).(string)
	defer sfs.closeTunnel(sessionId)

	if err, found := sfs.tunnelErrors[sessionId]; found {
		delete(sfs.tunnelErrors, sessionId)
		session.Error(err.Error())
		return
	}

	tunnel := sfs.getTunnel(sessionId)
	if tunnel == nil {
		session.Error("Ssh tunnel was not established")
		return
	}

	tunnel.session = session
	proxy := sfs.balancer.GetProxyBySessionId(sessionId)
	if proxy == nil {
		session.Error("Proxy was not found")
		return
	}

	tunnelEstablishedSshMessage(session, proxy, sfs.uiDomain)
	tunnel.conn.Wait()

	session.Error("Connection closed")
}

func (sfs *SSHForwardServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dest := sfs.balancer.GetProxyByAddress(r.Host)
	if nil == dest {
		ProxyNotFound(w)
		return
	}
	tunnel := dest.Tunnel
	if nil == tunnel {
		panic("tunnel not found while ssh-forwarding")
	}

	sfs.tracker.RequestStarted(r)

	originAddr, originPortStr, _ := net.SplitHostPort(r.RemoteAddr)
	originPort, _ := strconv.Atoi(originPortStr)

	ch, err := tunnel.GetChannel(dest.RequestedAddr, dest.Port, originAddr, uint32(originPort))
	if err != nil {
		RemoteServerNotFound(w)
		return
	}

	err = r.WriteProxy(ch)
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

func (sfs *SSHForwardServer) addTunnel(ctx ssh.Context, reqPayload remoteForwardRequest) error {
	tunnel := &SshTunnel{
		sessionId:    ctx.Value(ssh.ContextKeySessionID).(string),
		conn:         ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn),
		user:         ctx.Value(ssh.ContextKeyUser).(string),
		publicKey:    ctx.Value(ssh.ContextKeyPublicKey).(ssh.PublicKey),
		publicKeyStr: string(ctx.Value(ssh.ContextKeyPublicKey).(ssh.PublicKey).Marshal()),
	}

	_, err := sfs.balancer.CreateNewForward(reqPayload.BindAddr, reqPayload.BindPort, tunnel)
	if err != nil {
		sfs.tunnelErrors[tunnel.sessionId] = err
		return err
	}

	sfs.tunnels[tunnel.sessionId] = tunnel

	return nil
}

func (sfs *SSHForwardServer) getTunnel(id string) *SshTunnel {
	return sfs.tunnels[id]
}

func (sfs *SSHForwardServer) closeTunnel(id string) {
	if tunnel := sfs.getTunnel(id); nil != tunnel {
		sfs.balancer.DeleteProxyForSession(tunnel.sessionId)
		tunnel.CloseSession()
	}
	delete(sfs.tunnels, id)
}

type SshTunnel struct {
	sessionId    string
	conn         *gossh.ServerConn
	session      *Session
	user         string
	publicKey    ssh.PublicKey
	publicKeyStr string
}

func (tunnel *SshTunnel) GetChannel(destAddr string, destPort uint32, originAddr string, originPort uint32) (io.ReadWriteCloser, error) {
	payload := gossh.Marshal(&remoteForwardChannelData{
		DestAddr:   destAddr,
		DestPort:   destPort,
		OriginAddr: originAddr,
		OriginPort: originPort,
	})

	channel, reqs, err := tunnel.conn.OpenChannel(forwardedTCPChannelType, payload)
	if err != nil {
		return nil, err
	}
	go gossh.DiscardRequests(reqs)

	return channel, nil
}

func (tunnel *SshTunnel) CloseSession() {
	if nil != tunnel.session {
		_ = tunnel.session.Close()
	}
}

type Session struct {
	ssh.Session
}

func (session Session) Error(err string) {
	fmt.Fprintf(session, "\u001b[31m%s\u001B[0m\n", err)
}

func NewSshForwardServer(balancer *Balancer, tracker *client.TrafficTracker, port uint32, privateKey string, uiDomain string) *SSHForwardServer {
	return &SSHForwardServer{
		port:         port,
		privateKey:   privateKey,
		uiDomain:     uiDomain,
		balancer:     balancer,
		tracker:      tracker,
		tunnels:      make(map[string]*SshTunnel),
		tunnelErrors: make(map[string]error),
	}
}

func tunnelEstablishedSshMessage(session ssh.Session, proxy *Proxy, uiWebHost string) {
	fmt.Fprintf(session, "\u001b[32mYou proxy has been established:\u001b[0m\n")
	fmt.Fprintf(session, "Proxy:\t\t%s\n", proxy.Host())
	fmt.Fprintf(session, "Web ui:\t\thttp://%s/%s?token=%s\n\n", uiWebHost, proxy.Domain, proxy.Tunnel.sessionId)
}
