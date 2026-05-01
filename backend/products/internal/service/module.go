package service

import (
	"guru/utils/outbox"
	"guru/utils/pgsql"

	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(
		NewProductService,
		func(b *outbox.Builder) OutboxSaver { return b },
		func(tm *pgsql.TransactionManager) TxManager { return tm },
	),
)
