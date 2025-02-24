package chord

import "github.com/owlmq/owlmq/storage"

type Chord struct {
	storage     storage.StorageLayer
	fingertable []*FingerEntry
}

func New(s storage.StorageLayer) *Chord {
	return &Chord{
		storage: s,
	}
}
