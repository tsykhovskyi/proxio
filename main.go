package main

import (
	"fmt"
	"proxio/proxy"
	"proxio/ui"
)

func main() {
	var targetUrl = "http://localhost:8081"
	var proxyPortAccept = ":8001"
	var uiUrl = "http://127.0.0.1:4001"

	messagesChannel := proxy.ListenAndServe(proxyPortAccept, targetUrl)
	ui.Serve(uiUrl, messagesChannel)

	fmt.Printf("Forwarding: %s\t->\t%s\n", proxyPortAccept, targetUrl)
	fmt.Printf("Web interface: %s\n\n", uiUrl)

	// tunnel()

	select {}
}

// func publicKeyFile(file string) ssh.AuthMethod {
// 	buffer, err := ioutil.ReadFile(file)
// 	if err != nil {
// 		log.Fatalln(fmt.Sprintf("Cannot read SSH public key file %s", file))
// 		return nil
// 	}
//
// 	key, err := ssh.ParsePrivateKey(buffer)
// 	if err != nil {
// 		log.Fatalln(fmt.Sprintf("Cannot parse SSH public key file %s", file))
// 		return nil
// 	}
// 	return ssh.PublicKeys(key)
// }
//
// func tunnel() {
// 	sshConfig := &ssh.ClientConfig{
// 		// SSH connection username
// 		User: "operatore",
// 		Auth: []ssh.AuthMethod{
// 			// put here your private key path
// 			publicKeyFile("/Users/itsykhovskyi/.ssh/id_rsa"),
// 		},
// 		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
// 	}
//
// 	serverConn, err := ssh.Dial("tcp", "serveo.net:22", sshConfig)
// 	if err != nil {
// 		log.Fatalln(fmt.Printf("Dial INTO remote server error: %s", err))
// 	}
//
// 	listener, err := serverConn.Listen("tcp", ":80")
// 	if err != nil {
// 		log.Fatalln(fmt.Printf("Listen open port ON remote server error: %s", err))
// 	}
// 	defer listener.Close()
//
// 	for {
// 		// Open a (local) connection to localEndpoint whose content will be forwarded so serverEndpoint
// 		local, err := net.Dial("tcp", "localhost:8001")
// 		if err != nil {
// 			log.Fatalln(fmt.Printf("Dial INTO local service error: %s", err))
// 		}
//
// 		client, err := listener.Accept()
// 		if err != nil {
// 			log.Fatalln(err)
// 		}
//
// 		handleClient(client, local)
//
// 	}
// }
//
// func handleClient(client net.Conn, remote net.Conn) {
// 	defer client.Close()
// 	chDone := make(chan bool)
//
// 	// Start remote -> local data transfer
// 	go func() {
// 		_, err := io.Copy(client, remote)
// 		if err != nil {
// 			log.Println(fmt.Sprintf("error while copy remote->local: %s", err))
// 		}
// 		chDone <- true
// 	}()
//
// 	// Start local -> remote data transfer
// 	go func() {
// 		_, err := io.Copy(remote, client)
// 		if err != nil {
// 			log.Println(fmt.Sprintf("error while copy local->remote: %s", err))
// 		}
// 		chDone <- true
// 	}()
//
// 	<-chDone
// }
