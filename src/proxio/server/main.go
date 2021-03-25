package server

import (
	"os"
	"proxio/config"
	"proxio/event"
	"proxio/ssh"
	"proxio/ui"
	"sync"
)

var (
	wg  sync.WaitGroup
	err error
)

func Main() {
	configs := config.ParseApplicationArgs()
	uiDomain := "ui." + configs.Host

	suggester := ssh.NewCombinedSuggester()
	balancer := ssh.NewBalancer(configs.Host, suggester)
	sshServer := ssh.NewSshForwardServer(balancer, configs.SshPort, configs.PrivateKeyPath, uiDomain)
	uiHandler := ui.Handler(balancer)
	httpServer := ssh.NewHttpServer(sshServer, uiHandler, configs.Host, uiDomain)

	event.HandleTraffic()

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
