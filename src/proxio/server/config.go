package server

import "strconv"

type Configs struct {
	Host           string
	SshPort        uint32
	PrivateKeyPath string
}

type Config func(*Configs)

func Host(host string) Config {
	return func(configs *Configs) {
		configs.Host = host
	}
}

func Port(port uint32) Config {
	return func(configs *Configs) {
		configs.SshPort = port
	}
}

func PrivateKeyPath(keyPath string) Config {
	return func(configs *Configs) {
		configs.PrivateKeyPath = keyPath
	}
}

func NewConfig(setters ...Config) *Configs {
	configs := &Configs{
		SshPort: 22,
	}

	for _, configSet := range setters {
		configSet(configs)
	}

	return configs
}

type YamlConfig struct {
	Host           string `yaml:"host"`
	SshPort        string `yaml:"ssh_port"`
	PrivateKeyPath string `yaml:"private_key_path"`
}

func (yc *YamlConfig) ToConfig() (*Configs, error) {
	port, err := strconv.Atoi(yc.SshPort)
	if err != nil {
		return nil, err
	}

	return NewConfig(Host(yc.Host), Port(uint32(port)), PrivateKeyPath(yc.PrivateKeyPath)), nil
}
