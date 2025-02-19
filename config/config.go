package config

import (
	"math/big"
	"sync"

	"github.com/owlmq/owlmq/crypto"
)

type Config struct {
	Hostname    string
	NodeID      *big.Int
	Predecessor string
	Successor   string

	FingerTable []string

	//passwords
	Password_JOIN   string
	Password_PLUGIN string
}

var instance *Config
var once sync.Once

func New(hostname string) *Config {
	once.Do(func() {
		instance = &Config{
			Hostname: hostname,
			NodeID:   crypto.HashKey(hostname),
			//init Predecessor and Successor with itself
			Predecessor: hostname,
			Successor:   hostname,

			FingerTable: []string{},
			//generate initial passwords for nodes joining and passwords connecting
			Password_JOIN:   crypto.GenerateSecurePassword(16),
			Password_PLUGIN: crypto.GenerateSecurePassword(16),
		}
	})
	return instance
}

func GetInstance() *Config {
	return instance
}

func GetKnownNodes() []string {
	ret := []string{instance.Successor, instance.Predecessor}
	ret = append(ret, instance.FingerTable...)
	return ret
}
