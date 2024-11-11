package main

import (
	"time"

	"github.com/google/uuid"
)

// currentStore is able to store point current values
type currentStore interface {
	getCurrent(uuid.UUID) current
	setCurrent(uuid.UUID, currentInput)
}

type currentInput struct {
	Value *float64 `json:"value"`
}

type current struct {
	Ts    *time.Time `json:"ts"`
	Value *float64   `json:"value"`
}

// inMemoryCurrentStore stores point current values in a local in-memory cache.
// These are not shared between instances.
type inMemoryCurrentStore struct {
	cache map[uuid.UUID]current
}

func newInMemoryCurrentStore() inMemoryCurrentStore {
	return inMemoryCurrentStore{cache: map[uuid.UUID]current{}}
}

func (s inMemoryCurrentStore) getCurrent(id uuid.UUID) current {
	return s.cache[id]
}

func (s inMemoryCurrentStore) setCurrent(id uuid.UUID, input currentInput) {
	timestamp := time.Now()
	s.cache[id] = current{
		Ts:    &timestamp,
		Value: input.Value,
	}
}
