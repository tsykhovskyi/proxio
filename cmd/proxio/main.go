package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"proxio/client"
	"proxio/ui"
)

type Endpoint struct {
	Host string
	Port int
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", e.Host, e.Port)
}

func (e *Endpoint) Url() string {
	return fmt.Sprintf("http://%s:%d", e.Host, e.Port)
}

// local service to be forwarded
var localEndpoint = Endpoint{
	Host: "localhost",
	Port: 81,
}

// remote SSH server
var serverEndpoint = Endpoint{
	Host: "localhost",
	Port: 2222,
}

// remote forwarding port (on remote SSH server network)
var remoteEndpoint = Endpoint{
	Host: "localhost",
	Port: 83,
}

// web UI
var webUiEndpoint = Endpoint{
	Host: "localhost",
	Port: 4001,
}

func main() {
	l := tunnel(serverEndpoint, remoteEndpoint)
	defer l.Close()

	messagesChannel := client.ListenAndServe(l, localEndpoint.Url())
	ui.Serve(webUiEndpoint.String(), messagesChannel)

	fmt.Printf("Forwarding: %s\t->\t%s\n", remoteEndpoint.String(), localEndpoint.String())
	fmt.Printf("Web interface: %s\n\n", webUiEndpoint.String())

	select {}
}

func publicKeyFile(file string) ssh.AuthMethod {
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
	return ssh.PublicKeys(key)
}

type remoteForwardRequest struct {
	BindAddr string
	BindPort uint32
}

func tunnel(serverEndpoint, remoteEndpoint Endpoint) net.Listener {
	sshConfig := &ssh.ClientConfig{
		// SSH connection username
		User: "root",
		Auth: []ssh.AuthMethod{
			// put here your private key path
			publicKeyFile("/Users/itsykhovskyi/.ssh/id_digital_ocean"),
			// publicKeyFile("/Users/itsykhovskyi/.ssh/id_parkoss"),
			// ssh.KeyboardInteractive(SshInteractive),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		// HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// 	return nil
		// },
	}

	serverConn, err := ssh.Dial("tcp", serverEndpoint.String(), sshConfig)
	// serverConn, err := ssh.Dial("tcp", "serveo.net:22", sshConfig)
	if err != nil {
		log.Fatalln(fmt.Printf("Dial INTO remote server error: %s", err))
	}

	m := remoteForwardRequest{
		remoteEndpoint.Host,
		uint32(remoteEndpoint.Port),
	}
	// send message
	_, _, err = serverConn.SendRequest("prepare-tcpip-forward", true, ssh.Marshal(&m))
	if err != nil {
		log.Fatalln(fmt.Printf("Unable"))
	}

	listener, err := serverConn.Listen("tcp", remoteEndpoint.String())
	if err != nil {
		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
	}

	return listener
}
