package database

import "sync"

type PubSub struct {
	mu sync.RWMutex
	// map of channel name -> map of client channels
	subs map[string]map[chan string]struct{}
}

func NewPubSub() *PubSub {
	return &PubSub{
		subs: make(map[string]map[chan string]struct{}),
	}
}

func (ps *PubSub) Subscribe(topic string, clientChan chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.subs[topic]; !exists {
		ps.subs[topic] = make(map[chan string]struct{})
	}
	ps.subs[topic][clientChan] = struct{}{}
}

func (ps *PubSub) UnSubscribe(topic string, clientChan chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if _, exists := ps.subs[topic]; exists {
		if len(ps.subs[topic]) == 0 {
			delete(ps.subs, topic)
		}
	}
}

func (ps *PubSub) Publish(topic, message string) int {
	ps.mu.RLock()
	defer ps.mu.RLock()

	count := 0
	if subscribers, exists := ps.subs[topic]; exists {
		for clientChan := range subscribers {
			select {
			case clientChan <- message:
				count++
			default:
				// skip
			}
		}
	}

	return count
}
