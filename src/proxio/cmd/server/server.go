package main

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"proxio/server"
)

func main() {
	config := parseFlags()

	server.StartSSHServer(config)
}

func parseFlags() *server.Configs {
	configFilePath := flag.String("config", "", "path to config.yaml file")
	configAddress := flag.String("address", "", "server host")
	configPort := flag.Int("port", 22, "port for accepting incoming ssh connection")
	configPrivateKeyPath := flag.String("key", "", "path to ssh server private RSA key")

	flag.Parse()

	if *configFilePath != "" {
		data, err := ioutil.ReadFile(*configFilePath)
		if nil != err {
			panic("no configuration file")
		}
		yamlConfig := &server.YamlConfig{}
		err = yaml.Unmarshal(data, &yamlConfig)
		if nil != err {
			panic("wrong configuration. " + err.Error())
		}

		config, err := yamlConfig.ToConfig()
		if err != nil {
			panic(err)
		}

		return config
	}

	var setters []server.Config

	if *configAddress != "" {
		setters = append(setters, server.Host(*configAddress))
	}
	setters = append(setters, server.Port(uint32(*configPort)))
	if *configPrivateKeyPath != "" {
		setters = append(setters, server.PrivateKeyPath(*configPrivateKeyPath))
	}

	return server.NewConfig(setters...)
}
