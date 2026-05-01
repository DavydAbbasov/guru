package service

import (
	"guru/utils/idempotency"

	"go.uber.org/fx"
)

var Module = fx.Module("service",
	fx.Provide(
		NewNotificationService,
		func(r *idempotency.Repository) IdempotencyRepository { return r },
	),
)
