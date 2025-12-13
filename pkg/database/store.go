package database

import (
	"container/list"
	"fmt"
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

func (s *Store) HSet(key, field, value string, ttl time.Duration) (bool, error) {
	shard := s.getShard(key)
	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	item, exists := shard.Items[key]

	expiry := int64(0)
	if ttl > 0 {
		expiry = time.Now().Add(ttl).UnixNano()
	}

	if !exists {
		shard.Items[key] = &Item{
			Value:     map[string]string{field: value},
			Type:      TypeHash,
			ExpiresAt: expiry,
		}
		return true, nil
	}

	if item.Type != TypeHash {
		return false, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	hash := item.Value.(map[string]string)

	_, fieldExists := hash[field]
	hash[field] = value

	return !fieldExists, nil
}

func (s *Store) HGet(key, field string) (string, bool) {
	shard := s.getShard(key)

	shard.Mu.RLock()

	item, exists := shard.Items[key]
	if !exists {
		return "", false
	}

	// check if expired
	if item.ExpiresAt > 0 && time.Now().UnixNano() > item.ExpiresAt {
		// Delete item
		shard.Mu.RUnlock()
		s.Delete(key)
		return "", false
	}

	// Check type (If it's a String, you can't HGET it)
	if item.Type != TypeHash {
		return "", false
	}

	hash := item.Value.(map[string]string)
	val, ok := hash[field]

	shard.Mu.RUnlock()
	return val, ok
}

// LPush adds a value to the head of the list
// Returns the new length of the list
func (s *Store) LPush(key, value string, ttl time.Duration) (int, error) {
	shard := s.getShard(key)

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	expiry := int64(0)
	if ttl > 0 {
		expiry = time.Now().Add(ttl).UnixNano()
	}

	item, exists := shard.Items[key]
	if !exists {
		l := list.New()
		l.PushFront(value)
		shard.Items[key] = &Item{
			Value:     l,
			Type:      TypeList,
			ExpiresAt: expiry,
		}
		return 1, nil
	}

	if item.Type != TypeList {
		return 0, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	l := item.Value.(*list.List)
	l.PushFront(value)

	return l.Len(), nil
}

// LPop removes and returns the first element of the list
func (s *Store) LPop(key string) (string, bool) {
	shard := s.getShard(key)
	shard.Mu.RLock()
	defer shard.Mu.RUnlock()

	item, exists := shard.Items[key]
	if !exists {
		return "", false
	}

	if item.Type != TypeList {
		return "", false
	}

	l := item.Value.(*list.List)
	if l.Len() == 0 {
		return "", false
	}

	element := l.Front()
	val := element.Value.(string)

	l.Remove(element)

	if l.Len() == 0 {
		delete(shard.Items, key)
	}

	return val, true
}

func (s *Store) LRange(key string, start, stop int) ([]string, bool) {
	shard := s.getShard(key)
	shard.Mu.RLock()
	defer shard.Mu.RUnlock()

	item, exists := shard.Items[key]
	if !exists {
		return nil, false
	}

	if item.Type != TypeList {
		return nil, false
	}

	l := item.Value.(*list.List)
	length := l.Len()

	// handle negative
	if start < 0 {
		start = length + start
		if start < 0 {
			start = 0
		}
	}
	if stop < 0 {
		stop = length + stop
	}

	result := make([]string, 0, stop-start+1)

	current := l.Front()
	i := 0

	// skip until 'start'
	for i < start && current != nil {
		current = current.Next()
		i++
	}

	// collect until 'stop'
	for i < stop && current != nil {
		result = append(result, current.Value.(string))
		current = current.Next()
		i++
	}

	return result, true
}

func (s *Store) Delete(key string) {
	shard := s.getShard(key)

	shard.Mu.Lock()
	defer shard.Mu.Unlock()

	delete(shard.Items, key)
}
