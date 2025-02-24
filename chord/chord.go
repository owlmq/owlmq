package chord

import "github.com/owlmq/owlmq/storage"

type Chord struct {
	storage storage.StorageLayer
}

func New(s storage.StorageLayer) *Chord {
	return &Chord{
		storage: s,
	}
}
