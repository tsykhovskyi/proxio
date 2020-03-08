package server

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"log"
	"strconv"
)

func StartSSHServer(configs *Configs) {
	handler := func(s ssh.Session) {
		key := gossh.MarshalAuthorizedKey(s.PublicKey())
		out := fmt.Sprintf("Hi, %s\n", key)
		io.WriteString(s, out)
	}

	s := &ssh.Server{
		Addr:    ":" + strconv.Itoa(int(configs.SshPort)),
		Handler: handler,
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
	s.AddHostKey(publicKeyFile(configs.PrivateKeyPath))

	servers := NewForwardServers()

	balancer := NewBalancer(servers)
	s.RequestHandlers = map[string]ssh.RequestHandler{
		"prepare-tcpip-forward": balancer.HandleSSHRequest, // custom type, sending from client application
		"tcpip-forward":         balancer.HandleSSHRequest,
		"cancel-tcpip-forward":  balancer.HandleSSHRequest,
	}

	s.ChannelHandlers = ssh.DefaultChannelHandlers
	s.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	log.Fatal(s.ListenAndServe())
}
