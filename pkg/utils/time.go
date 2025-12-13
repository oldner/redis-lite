package utils

import "time"

func ParseDuration(arg string) time.Duration {
	var ttl time.Duration
	if d, err := time.ParseDuration(arg); err == nil {
		ttl = d
	}

	return ttl
}
