package server

import (
	"net/http"
	"os"
	"proxio"
	"proxio/client"
	"proxio/config"
	"proxio/ui"
	"sync"
)

var locator = proxio.NewLocator()

func Main() {
	configs := config.ParseApplicationArgs()

	locator.Add("balancer", NewBalancer())
	locator.Add("ssh-forward-server", NewSshForwardServer(locator.Get("balancer").(*Balancer), configs.SshPort, configs.PrivateKeyPath))
	locator.Add("traffic-tracker", client.NewTrafficTracker(locator.Get("balancer").(*Balancer)))
	locator.Add("ui", ui.Handler(locator.Get("traffic-tracker").(*client.TrafficTracker)))
	locator.Add("http-server", NewHttpServer(locator.Get("traffic-tracker").(*client.TrafficTracker), locator.Get("ui").(http.Handler)))

	sshServer := locator.Get("ssh-forward-server").(*SSHForwardServer)
	httpServer := locator.Get("http-server").(*http.Server)

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
