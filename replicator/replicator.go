package replicator

import (
	"errors"

	"github.com/owlmq/owlmq/chord"
	"github.com/owlmq/owlmq/storage"
)

type Replicator interface {
	Start()
}

func New(rt ReplicatorType, s storage.StorageLayer, c chord.Chord) (Replicator, error) {
	switch rt {
	case SuccessorList:
		return newSuccessorListReplicator(), nil
	case VirtualNode:
		return nil, errors.New("Unimplemented Storagelayer Type")
	default:
		return nil, errors.New("Unknown StorageLayer Type")
	}
}
