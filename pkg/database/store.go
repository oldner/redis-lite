package database

import (
	"sync"
	"time"
)

// DataType helps us manage different Redis types (String, List, Set, etc.)
type DataType int

const (
	TypeString DataType = iota
	TypeList
	TypeSet
	TypeHash
)

// Item represents the value stored in memory.
// It holds the actual data and metadata like expiration.
type Item struct {
	Value     interface{} // Can be string, []string, map[string]string, etc.
	Type      DataType
	ExpiresAt int64 // Unix timestamp (0 means no expiration)
}

// Store is the main database struct.
type Store struct {
	mu   sync.RWMutex     // Reader-Writer Lock for thread safety
	data map[string]*Item // The core storage
}

// NewStore initializes the DB.
func NewStore() *Store {
	return &Store{
		data: make(map[string]*Item),
	}
}

func (s *Store) Set(key string, value interface{}, ttl time.Duration) {
	s.mu.Lock() // LOCK for writing
	defer s.mu.Unlock()

	expiry := int64(0)
	if ttl > 0 {
		expiry = time.Now().Add(ttl).UnixNano()
	}

	s.data[key] = &Item{
		Value:     value,
		Type:      TypeString,
		ExpiresAt: expiry,
	}
}

func (s *Store) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.Unlock()

	item, exists := s.data[key]
	if !exists {
		return nil, false
	}

	if item.ExpiresAt > 0 && time.Now().UnixNano() > item.ExpiresAt {
		// Delete item
		s.Delete(key)
		return nil, false
	}

	return item.Value, true
}

func (s *Store) Delete(key string) {
	s.mu.RLock()
	defer s.mu.Unlock()

	delete(s.data, key)
}
