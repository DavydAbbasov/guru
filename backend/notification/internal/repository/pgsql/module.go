package pgsql

import (
	"guru/backend/notification/internal/repository"

	"go.uber.org/fx"
)

var Module = fx.Module("notification_pgsql_repository",
	fx.Provide(
		fx.Annotate(
			NewEventRepository,
			fx.As(new(repository.EventRepository)),
		),
	),
)
