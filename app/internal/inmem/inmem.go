package inmem

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	mu    sync.Mutex
	store map[string]any
}

func (s *Store) Get(key string) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.store[key]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (s *Store) Set(key string, val any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = val
	return nil
}
