package idempotency

import "time"

type Config struct {
	Namespace       string
	Retention       time.Duration
	CleanupInterval time.Duration
}

func (c *Config) withDefaults() {
	if c.Retention <= 0 {
		c.Retention = 7 * 24 * time.Hour
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = time.Hour
	}
}
