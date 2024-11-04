package main

import (
	"time"

	"github.com/google/uuid"
)

// currentStore is able to store point current values
type currentStore interface {
	getCurrent(uuid.UUID) apiCurrent
	setCurrent(uuid.UUID, apiCurrentInput)
}

// inMemoryCurrentStore stores point current values in a local in-memory cache.
// These are not shared between instances.
type inMemoryCurrentStore struct {
	cache map[uuid.UUID]apiCurrent
}

func newInMemoryCurrentStore() inMemoryCurrentStore {
	return inMemoryCurrentStore{cache: map[uuid.UUID]apiCurrent{}}
}

func (s inMemoryCurrentStore) getCurrent(id uuid.UUID) apiCurrent {
	return s.cache[id]
}

func (s inMemoryCurrentStore) setCurrent(id uuid.UUID, input apiCurrentInput) {
	timestamp := time.Now()
	s.cache[id] = apiCurrent{
		Ts:    &timestamp,
		Value: input.Value,
	}
}
