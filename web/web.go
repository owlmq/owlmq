package web

import (
	"sync"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

type OwlmqServer struct {
	pb.UnimplementedOwlmqServer
	store map[string]string
	mu    sync.RWMutex
}

func NewOwlmqServer() *OwlmqServer {
	return &OwlmqServer{
		store: make(map[string]string),
	}
}
