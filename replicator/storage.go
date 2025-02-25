package replicator

import (
	"errors"
	"sync"
)

func newReplicaStorage() *ReplicaStorage {
	return &ReplicaStorage{
		store: make(map[string]ReplicatedEntry),
	}
}

type ReplicaStorage struct {
	store map[string]ReplicatedEntry
	mu    sync.RWMutex
}

func (s *ReplicaStorage) Put(key string, value ReplicatedEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = value
}

func (s *ReplicaStorage) Get(key string) (*ReplicatedEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.store[key]
	if !exists {
		return nil, errors.New("Key not found")
	}

	return &value, nil
}
