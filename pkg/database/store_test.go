package database

import (
	"redis-lite/pkg/utils"
	"sync"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	s := NewStore()
	key := "foo"
	val := "bar"

	// 1. Test Set
	s.Set(key, val, 0)

	// 2. Test Get
	got, found := s.Get(key)
	if !found {
		t.Fatalf("Expected key %s to exist", key)
	}

	if got != val {
		t.Errorf("Expected %v, got %v", val, got)
	}
}

func TestHSetHGet(t *testing.T) {
	s := NewStore()
	key := "foo"
	field := "foofield"
	val := "bar"
	expiry := "1m"

	ttl := utils.ParseDuration(expiry)

	// 1. Test HSet
	s.HSet(key, field, val, ttl)

	// 2. Test Get
	got, found := s.HGet(key, field)
	if !found {
		t.Fatalf("Expected key %s or field %s to exist", key, field)
	}

	if got != val {
		t.Errorf("Expected %v, got %v", val, got)
	}
}

func TestLPushLPopLRange(t *testing.T) {
	s := NewStore()
	key := "foo"
	val := "bar"
	expiry := "1m"

	ttl := utils.ParseDuration(expiry)

	// 1. Test LPush
	s.LPush(key, val, ttl)

	// 2. Test LPop
	got, found := s.LPop(key)
	if !found {
		t.Fatalf("Expected key %s to exist", key)
	}

	if got != val {
		t.Errorf("Expected %v, got %v", val, got)
	}

	// 3. Test LRange
	s.LPush(key, val, ttl)
	s.LPush(key, val, ttl)

	list, ok := s.LRange(key, 0, 2)
	if !ok {
		t.Errorf("Expected %v, got %v", val, got)
	}

	if len(list) != 2 {
		t.Errorf("Expected 2 values but got %d", len(list))
	}
}

func TestExpiration(t *testing.T) {
	s := NewStore()
	key := "shortlived"
	val := "data"

	// Set with 10ms TTL
	s.Set(key, val, 10*time.Millisecond)

	// Immediate check (Should exist)
	_, found := s.Get(key)
	if !found {
		t.Fatal("Key should exist immediately after set")
	}

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Check again (Should be gone)
	_, found = s.Get(key)
	if found {
		t.Error("Key should have expired")
	}
}

// It fires 100 goroutines to read/write simultaneously.
// Run with: go test -race
func TestConcurrency(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup

	// Write 100 keys concurrently
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s.Set("key", "value", 0)
		}(i)
	}

	// Read concurrently while writing
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Get("key")
		}()
	}

	wg.Wait()
	// If the test finishes without panicking or race conditions, we pass.
}
