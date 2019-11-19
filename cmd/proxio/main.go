package main

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"proxio/proxy"
	"proxio/ui"
)

func main() {
	var targetUrl = "http://localhost:8011"
	var uiUrl = "http://127.0.0.1:4001"

	l, _ := net.Listen("tcp", ":8012")
	// l := tunnel()
	defer l.Close()

	messagesChannel := proxy.ListenAndServe(l, targetUrl)
	ui.Serve(uiUrl, messagesChannel)

	fmt.Printf("Forwarding: %s\t->\t%s\n", "", targetUrl)
	fmt.Printf("Web interface: %s\n\n", uiUrl)

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

func tunnel() net.Listener {
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

	serverConn, err := ssh.Dial("tcp", "68.183.216.91:2329", sshConfig)
	// serverConn, err := ssh.Dial("tcp", "serveo.net:22", sshConfig)
	if err != nil {
		log.Fatalln(fmt.Printf("Dial INTO remote server error: %s", err))
	}

	listener, err := serverConn.Listen("tcp", "0.0.0.0:55012")
	if err != nil {
		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
	}

	return listener
}
