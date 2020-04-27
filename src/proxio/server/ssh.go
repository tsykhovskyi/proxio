package server

import (
	"bufio"
	"errors"
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
	balancer     *Balancer
	tracker      *client.TrafficTracker
	tunnels      map[string]*SshTunnel
	tunnelErrors map[string]string
	port         uint32
	privateKey   string
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
			// TODO: log parse failure
			return false, []byte{}
		}
		err := sfs.addTunnel(ctx, reqPayload)
		if err != nil {
			return false, nil
		}

		return true, gossh.Marshal(&remoteForwardSuccess{reqPayload.BindPort})

	// case "cancel-tcpip-forward":
	// 	var reqPayload remoteForwardCancelRequest
	// 	if err := gossh.Unmarshal(req.Payload, &reqPayload); err != nil {
	// 		// TODO: log parse failure
	// 		return false, []byte{}
	// 	}
	//
	// 	return true, nil
	default:
		return false, nil
	}
}

func (sfs *SSHForwardServer) HandleSshSession(session ssh.Session) {
	_ = gossh.MarshalAuthorizedKey(session.PublicKey())
	sessionId := session.Context().Value(ssh.ContextKeySessionID).(string)
	if err, found := sfs.tunnelErrors[sessionId]; found {
		delete(sfs.tunnelErrors, sessionId)
		session.Write([]byte(err + "\n"))
		session.Close()
		return
	}

	tunnel := sfs.getTunnel(sessionId)
	if tunnel == nil {
		session.Write([]byte("Proxy was not established\n"))
		return
	}

	tunnel.session = session

	tunnel.conn.Wait()

	sfs.closeTunnel(sessionId)
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
		sessionId: ctx.Value(ssh.ContextKeySessionID).(string),
		conn:      ctx.Value(ssh.ContextKeyConn).(*gossh.ServerConn),
		user:      ctx.Value(ssh.ContextKeyUser).(string),
		publicKey: ctx.Value(ssh.ContextKeyPublicKey).(ssh.PublicKey),
	}

	if _, err := sfs.balancer.ValidateRequestDomain(reqPayload.BindAddr, reqPayload.BindPort); err != "" {
		sfs.tunnelErrors[tunnel.sessionId] = err
		return errors.New(err)
	}

	domain := sfs.balancer.CreateNewForward(reqPayload.BindAddr, reqPayload.BindPort, tunnel)
	fmt.Println("generated domain is " + domain)

	sfs.tunnels[tunnel.sessionId] = tunnel

	return nil
}

func (sfs *SSHForwardServer) getTunnel(id string) *SshTunnel {
	return sfs.tunnels[id]
}

func (sfs *SSHForwardServer) closeTunnel(id string) {
	if tunnel := sfs.getTunnel(id); nil != tunnel {
		sfs.balancer.DeleteForwardForSession(tunnel.sessionId)
		tunnel.CloseSession()
	}
	delete(sfs.tunnels, id)
}

type SshTunnel struct {
	sessionId string
	conn      *gossh.ServerConn
	session   ssh.Session
	user      string
	publicKey ssh.PublicKey
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

func NewSshForwardServer(balancer *Balancer, tracker *client.TrafficTracker, port uint32, privateKey string) *SSHForwardServer {
	return &SSHForwardServer{
		balancer:     balancer,
		port:         port,
		privateKey:   privateKey,
		tracker:      tracker,
		tunnels:      make(map[string]*SshTunnel),
		tunnelErrors: make(map[string]string),
	}
}
