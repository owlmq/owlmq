package storage

import (
	"errors"
	"sync"
)

func newInMemoryStorage() StorageLayer {
	return &InMemoryStorage{
		store: make(map[string]string),
	}
}

type InMemoryStorage struct {
	store map[string]string
	mu    sync.RWMutex
}

func (s *InMemoryStorage) Put(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[key] = value
}

func (s *InMemoryStorage) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.store[key]
	if !exists {
		return "", errors.New("Key not found")
	}

	return value, nil
}

func (s *InMemoryStorage) Iterator() Iterator {
	return newMapIterator(s.store)
}
