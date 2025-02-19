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

func (c *Chord) Put(key string, value string) error {
	//TODO check if i am the correct node
	c.storage.Put(key, value)
	return nil
}

func (c *Chord) Get(key string) (string, error) {
	//TODO check if i am the correct node
	value, err := c.storage.Get(key)
	return value, err
}
