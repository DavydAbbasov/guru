package pgsql

import (
	"time"

	"guru/utils/pgsql"
)

var Module = pgsql.ModuleWithOptions(
	pgsql.WithLightPool(),
	pgsql.WithCommitTimeout(15*time.Second),
)
