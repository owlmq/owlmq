package storage

import "errors"

type StorageLayer interface {
	Put(key string, value string)
	Get(key string) (string, error)

	//interator over the storage
	Iterator() Iterator
}

type Iterator interface {
	Next() bool
	Value() (string, string)
}

func New(st StorageType) (StorageLayer, error) {
	switch st {
	case InMemory:
		return newInMemoryStorage(), nil
	case File:
		return nil, errors.New("Unimplemented Storagelayer Type")
	default:
		return nil, errors.New("Unknown StorageLayer Type")
	}
}
