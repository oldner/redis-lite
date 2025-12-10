package database

import (
	"log/slog"
	"redis-lite/pkg/cfg"
	"time"
)

type Janitor struct {
	Interval time.Duration
	stop     chan struct{}
}

func NewJanitor(config *cfg.Config) *Janitor {
	return &Janitor{
		Interval: config.JanitorInterval,
		stop:     make(chan struct{}),
	}
}

func (j *Janitor) Run(target *Store) {
	ticker := time.NewTicker(j.Interval)
	slog.Warn("Starting janitor ticker", "the interval of ", j.Interval.String())

	for {
		select {
		case <-ticker.C:
			j.vacuum(target)
		case <-j.stop:
			j.Stop()
			return
		}
	}
}

func (j *Janitor) vacuum(s *Store) {
	now := time.Now().UnixNano()
	for _, shard := range s.Shards {
		shard.Mu.Lock()
		for key, item := range shard.Items {
			if item.ExpiresAt > 0 && now > item.ExpiresAt {
				delete(shard.Items, key)
			}
		}
		shard.Mu.Unlock()
	}
}

func (j *Janitor) Stop() {
	slog.Warn("Stopping the scheduler..")
	close(j.stop)
}
