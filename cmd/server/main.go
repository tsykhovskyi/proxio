package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"proxio/server"
)

type ServerConfig struct {
	Host           string `yaml:"host"`
	SshPort        string `yaml:"ssh_port"`
	PrivateKeyPath string `yaml:"private_key_path"`
}

var (
	help           = flag.Bool("h", false, "print help")
	configFilePath = flag.String("config", "./config.yaml", "path to config.yaml file")
	config         *ServerConfig
)

func init() {
	flag.Parse()
	if *help {
		helpText()
		os.Exit(0)
	}

	data, err := ioutil.ReadFile(*configFilePath)
	if nil != err {
		panic("no configuration file")
	}
	config = &ServerConfig{}
	err = yaml.Unmarshal(data, &config)
	if nil != err {
		panic("wrong configuration. " + err.Error())
	}
}

func helpText() {
	fmt.Println("USAGE")
	fmt.Println("  server [flags]")
	fmt.Println("")
	fmt.Println("FLAGS")
	flag.PrintDefaults()
}

func main() {
	server.StartSSHServer(config.SshPort, config.PrivateKeyPath)
}
