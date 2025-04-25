package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

var C *Config

type Config struct {
	CertPath   string `yaml:"certPath"`
	KeyPath    string `yaml:"keyPath"`
	LocalHost  string `yaml:"localHost"`
	LocalPort  int    `yaml:"localPort"`
	RemoteHost string `yaml:"remoteHost"`
	RemotePort int    `yaml:"remotePort"`
	Password   string `yaml:"password"`
	ClientID   string `yaml:"clientID"`
}

func MustLoad(path string) {
	var config Config
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
	}
	C = &config
}
