package config

import (
	"math/big"
	"sync"

	"github.com/owlmq/owlmq/crypto"
)

// TODO later make this part of the configuration file
var M int64 = 32 //-> 10.000 nodes
//var M int64 = 64 //-> 1 million nodes
//var M int64 = 128  //-> 1 billion nodes
//var M int64 = 160  //-> realy big --

// TODO later make this part of the configuration file
// this is the len of the SuccessorList so the amound of Successors which are stored in it
var SuccessorListSize int = 4

type FingerEntry struct {
	Start   *big.Int
	Address string
}

type Config struct {
	Hostname    string
	NodeID      *big.Int
	Predecessor string
	Successor   string

	SuccessorList []string
	FingerTable   []*FingerEntry

	//passwords
	Password_JOIN   string
	Password_PLUGIN string
}

var instance *Config
var once sync.Once

func New(hostname string) *Config {
	once.Do(func() {
		instance = &Config{
			Hostname:    hostname,
			NodeID:      crypto.HashKey(hostname),
			Predecessor: hostname,
			Successor:   hostname,

			SuccessorList: make([]string, SuccessorListSize),
			FingerTable:   make([]*FingerEntry, M),
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
	for _, v := range instance.FingerTable {
		ret = append(ret, v.Address)
	}
	return ret
}

func SetSuccessor(address string) {
	instance.Successor = address
}

func SetPredecessor(address string) {
	instance.Predecessor = address
}
