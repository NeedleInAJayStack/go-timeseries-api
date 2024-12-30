package main

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// currentStore is able to store point current values
type currentStore interface {
	getCurrent(uuid.UUID) (current, error)
	setCurrent(uuid.UUID, currentInput) error
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
	mux   *sync.Mutex
	cache map[uuid.UUID]current
}

func newInMemoryCurrentStore() inMemoryCurrentStore {
	return inMemoryCurrentStore{
		mux:   &sync.Mutex{},
		cache: map[uuid.UUID]current{},
	}
}

func (s inMemoryCurrentStore) getCurrent(id uuid.UUID) (current, error) {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.cache[id], nil
}

func (s inMemoryCurrentStore) setCurrent(id uuid.UUID, input currentInput) error {
	timestamp := time.Now()
	s.mux.Lock()
	defer s.mux.Unlock()
	s.cache[id] = current{
		Ts:    &timestamp,
		Value: input.Value,
	}
	return nil
}

// redisCurrentStore stores point current values in a Redis database.
type redisCurrentStore struct {
	db        *redis.Client
	keyPrefix string
	ctx       context.Context
}

func newRedisCurrentStore(db *redis.Client) redisCurrentStore {
	return redisCurrentStore{
		db:        db,
		keyPrefix: "timeseries-api:",
		ctx:       context.Background(),
	}
}

func (s redisCurrentStore) getCurrent(id uuid.UUID) (current, error) {
	currentJson, err := s.db.Get(s.ctx, s.keyPrefix+id.String()).Bytes()
	if err != nil {
		log.Printf("Cannot retrieve current: %s", err)
		return current{}, err
	}
	var result current
	err = json.Unmarshal(currentJson, &result)
	if err != nil {
		log.Printf("Cannot decode current JSON: %s", err)
		return current{}, err
	}
	return result, nil
}

func (s redisCurrentStore) setCurrent(id uuid.UUID, input currentInput) error {
	timestamp := time.Now()
	currentJson, err := json.Marshal(current{
		Ts:    &timestamp,
		Value: input.Value,
	})
	if err != nil {
		log.Printf("Cannot encode current JSON: %s", err)
		return err
	}
	err = s.db.Set(s.ctx, s.keyPrefix+id.String(), currentJson, 0).Err()
	if err != nil {
		log.Printf("Cannot store current: %s", err)
		return err
	}
	return nil
}
