package server

import (
	"os"
	"proxio/client"
	"proxio/config"
	"proxio/ui"
	"sync"
)

func Main() {
	configs := config.ParseApplicationArgs()

	balancer := NewBalancer()
	sshServer := NewSshForwardServer(balancer, configs.SshPort, configs.PrivateKeyPath)
	httpTrafficHandler := client.NewTrafficTracker(sshServer)
	uiHandler := ui.Handler(httpTrafficHandler.GetTraffic())
	httpServer := NewHttpServer(httpTrafficHandler, uiHandler)

	var (
		wg  sync.WaitGroup
		err error
	)
	wg.Add(1)
	go func() {
		err = sshServer.Start()
		wg.Done()
	}()
	go func() {
		err = httpServer.ListenAndServe()
		wg.Done()
	}()

	wg.Wait()
	if err != nil {
		panic(err)
	}
	os.Exit(0)
}
