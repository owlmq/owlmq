package replicator

import "github.com/owlmq/owlmq/storage"

func newSuccessorListReplicator() Replicator {
	return &SuccessorListReplicator{}
}

type SuccessorListReplicator struct {
	storage storage.StorageLayer
}

func (s *SuccessorListReplicator) Start() {
	//TODO
	panic("Unimplemented")
}
