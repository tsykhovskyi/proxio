package server

import (
	"bufio"
	"fmt"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
)

func StartSSHServer(port string, serverKeyPath string) {
	handler := func(s ssh.Session) {
		key := gossh.MarshalAuthorizedKey(s.PublicKey())
		out := fmt.Sprintf("Hi, %s\n", key)
		io.WriteString(s, out)

		scanner := bufio.NewScanner(s)
		for scanner.Scan() {
			// fmt.Fprintln(s, "got:", scanner.Text())
		}
	}

	s := &ssh.Server{
		Addr:    ":" + port,
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
	s.AddHostKey(publicKeyFile(serverKeyPath))

	balancer := &Balancer{}
	s.RequestHandlers = map[string]ssh.RequestHandler{
		"tcpip-forward":        balancer.HandleSSHRequest,
		"cancel-tcpip-forward": balancer.HandleSSHRequest,
	}

	s.ChannelHandlers = ssh.DefaultChannelHandlers
	s.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	log.Fatal(s.ListenAndServe())
}

func publicKeyFile(file string) gossh.Signer {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH private key file %s", file))
		return nil
	}

	key, err := gossh.ParsePrivateKey(buffer)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}

	return key
}
