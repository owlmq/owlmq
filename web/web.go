package web

import (
	"sync"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
)

type OwlmqServer struct {
	pb.UnimplementedOwlmqServer
	chordLayer *chord.Chord
	mu         sync.RWMutex
}

func NewOwlmqServer(cl *chord.Chord) *OwlmqServer {
	return &OwlmqServer{
		chordLayer: cl,
	}
}
