package main

import (
	"bufio"
	"fmt"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
)

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

func main() {
	handler := func(s ssh.Session) {
		io.WriteString(s, "Hello world\n")

		scanner := bufio.NewScanner(s)
		for scanner.Scan() {
			fmt.Fprintln(s, "got:", scanner.Text())
		}

	}

	s := &ssh.Server{
		Addr:    ":2222",
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
	}
	s.AddHostKey(publicKeyFile("cmd/server/keys/id_rsa"))

	tcpIpForwardHandler := &ssh.ForwardedTCPHandler{}
	s.RequestHandlers = map[string]ssh.RequestHandler{
		"tcpip-forward":        tcpIpForwardHandler.HandleSSHRequest,
		"cancel-tcpip-forward": tcpIpForwardHandler.HandleSSHRequest,
	}

	s.ChannelHandlers = ssh.DefaultChannelHandlers
	s.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	log.Fatal(s.ListenAndServe())
}
