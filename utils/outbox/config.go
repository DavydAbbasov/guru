package outbox

import "time"

type Config struct {
	Namespace       string
	BatchSize       int
	PollInterval    time.Duration
	CleanupInterval time.Duration
	Retention       time.Duration
	MaxAttempts     int
	RetryBaseDelay  time.Duration
}

func (c *Config) withDefaults() {
	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}
	if c.PollInterval <= 0 {
		c.PollInterval = time.Second
	}
	if c.CleanupInterval <= 0 {
		c.CleanupInterval = time.Hour
	}
	if c.Retention <= 0 {
		c.Retention = 24 * time.Hour
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 10
	}
	if c.RetryBaseDelay <= 0 {
		c.RetryBaseDelay = 5 * time.Second
	}
}
