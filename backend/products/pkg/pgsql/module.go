package pgsql

import (
	"time"

	"guru/utils/pgsql"
)

var Module = pgsql.ModuleWithOptions(
	pgsql.WithPoolConfig(&pgsql.ConnectionPoolConfig{
		MaxIdleTime:            30 * time.Minute,
		ConnectionMaxLifetime:  time.Hour,
		MaxIdleConnectionCount: 5,
		MaxOpenConnectionCount: 25,
		ConnectionTimeout:      30 * time.Second,
	}),
	pgsql.WithCommitTimeout(15*time.Second),
)
