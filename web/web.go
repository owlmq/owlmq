package web

import (
	"sync"

	pb "github.com/owlmq/owlmq/api/owlmq"
	"github.com/owlmq/owlmq/chord"
	"github.com/owlmq/owlmq/messaging"
)

type OwlmqServer struct {
	pb.UnimplementedOwlmqServer
	chordLayer     *chord.Chord
	mu             sync.RWMutex
	messagingLayer messaging.MessagingLayer
}

func NewOwlmqServer(cl *chord.Chord, ml messaging.MessagingLayer) *OwlmqServer {
	return &OwlmqServer{
		chordLayer:     cl,
		messagingLayer: ml,
	}
}
