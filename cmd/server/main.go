package main

import (
	"fmt"
	server "github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
)

func publicKeyFile(file string) ssh.Signer {
	buffer, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot read SSH private key file %s", file))
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
		return nil
	}

	return key
}

func main() {
	handler := func(s server.Session) {
		io.WriteString(s, "Hello world\n")
	}

	s := &server.Server{
		Addr:    ":2222",
		Handler: handler,
		ConnCallback: func(conn net.Conn) net.Conn {
			return conn
		},
		LocalPortForwardingCallback: func(ctx server.Context, destinationHost string, destinationPort uint32) bool {
			return true
		},
		SessionRequestCallback: func(sess server.Session, requestType string) bool {
			return true
		},
		ReversePortForwardingCallback: func(ctx server.Context, bindHost string, bindPort uint32) bool {
			return true
		},
	}
	s.AddHostKey(publicKeyFile("cmd/server/keys/id_rsa"))

	log.Fatal(s.ListenAndServe())
}
