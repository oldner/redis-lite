package database

import (
	"hash/fnv"
	"sync"
	"time"
)

// 256 shards is a good balance for memory vs concurrency
const ShardCount = 256

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
	Value     interface{}
	Type      DataType
	ExpiresAt int64
}

type Shard struct {
	Mu    sync.RWMutex
	Items map[string]*Item
}

// Store is the main database struct.
type Store struct {
	Shards []*Shard
}

// NewStore initializes the DB.
func NewStore() *Store {
	s := &Store{
		Shards: make([]*Shard, ShardCount),
	}
	for i := 0; i < ShardCount; i++ {
		s.Shards[i] = &Shard{
			Items: make(map[string]*Item),
		}
	}
	return s
}

// getShardIndex hashes the key to find which shard it belongs to
func (s *Store) getShardIndex(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32()) % ShardCount
}

// getShard is a helper to retrieve the specific shard for a key
func (s *Store) getShard(key string) *Shard {
	return s.Shards[s.getShardIndex(key)]
}

func (s *Store) Set(key string, value interface{}, ttl time.Duration) {
	shard := s.getShard(key)

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	expiry := int64(0)
	if ttl > 0 {
		expiry = time.Now().Add(ttl).UnixNano()
	}

	shard.Items[key] = &Item{
		Value:     value,
		Type:      TypeString,
		ExpiresAt: expiry,
	}
}

func (s *Store) Get(key string) (interface{}, bool) {
	shard := s.getShard(key)

	shard.Mu.RLock()

	item, exists := shard.Items[key]
	if !exists {
		return nil, false
	}

	// check if expired
	if item.ExpiresAt > 0 && time.Now().UnixNano() > item.ExpiresAt {
		// Delete item
		shard.Mu.RUnlock()
		s.Delete(key)
		return nil, false
	}

	shard.Mu.RUnlock()
	return item.Value, true
}

func (s *Store) Delete(key string) {
	shard := s.getShard(key)

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	delete(shard.Items, key)
}
