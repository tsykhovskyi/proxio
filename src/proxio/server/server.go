package server

import (
	"log"
)

func Start(configs *Configs) {
	balancer := NewBalancer()

	forwardingHandler := NewSshForwardHandler(balancer)

	err := forwardingHandler.Start(configs.SshPort, configs.PrivateKeyPath)

	log.Fatal(err)
}
