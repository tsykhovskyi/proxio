package server

import (
	"os"
	"proxio/client"
	"proxio/config"
	"proxio/repository"
	"proxio/ssh"
	"proxio/ui"
	"sync"
)

func Main() {
	configs := config.ParseApplicationArgs()
	uiDomain := "ui." + configs.Host

	sessionRepo := repository.NewSessionsRepo()
	sessionRepo.Populate()

	suggester := ssh.NewCombinedSuggester()
	balancer := ssh.NewBalancer(configs.Host, suggester)
	httpTrafficHandler := client.NewTrafficTracker()
	sshServer := ssh.NewSshForwardServer(balancer, httpTrafficHandler, configs.SshPort, configs.PrivateKeyPath, uiDomain)
	uiHandler := ui.Handler(httpTrafficHandler.GetTraffic(), sessionRepo, balancer)
	httpServer := ssh.NewHttpServer(sshServer, uiHandler, configs.Host, uiDomain)

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
