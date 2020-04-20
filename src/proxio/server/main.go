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

	balancer := NewBalancer(configs.Host)
	httpTrafficHandler := client.NewTrafficTracker()
	sshServer := NewSshForwardServer(balancer, httpTrafficHandler, configs.SshPort, configs.PrivateKeyPath)
	uiHandler := ui.Handler(httpTrafficHandler.GetTraffic())
	httpServer := NewHttpServer(sshServer, uiHandler)

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
