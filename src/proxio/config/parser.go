package config

import (
	"flag"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func ParseApplicationArgs() *Configs {
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
		yamlConfig := &YamlConfig{}
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

	var setters []Config

	if *configAddress != "" {
		setters = append(setters, Host(*configAddress))
	}
	setters = append(setters, Port(uint32(*configPort)))
	if *configPrivateKeyPath != "" {
		setters = append(setters, PrivateKeyPath(*configPrivateKeyPath))
	}

	return NewConfig(setters...)
}
