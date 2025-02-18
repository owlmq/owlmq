package config

import (
	"github.com/owlmq/owlmq/crypto"
)

type Config struct {
	Hostname    string
	NodeID      string
	Predecessor string
	Successor   string

	//passwords
	Password_JOIN   string
	Password_PLUGIN string
}

func New(hostname string) *Config {
	return &Config{
		Hostname:    hostname,
		NodeID:      crypto.GenerateSHA1(hostname),
		Predecessor: "",
		Successor:   "",
		//generate initial passwords for nodes joining and passwords connecting
		Password_JOIN:   crypto.GenerateSecurePassword(16),
		Password_PLUGIN: crypto.GenerateSecurePassword(16),
	}
}
