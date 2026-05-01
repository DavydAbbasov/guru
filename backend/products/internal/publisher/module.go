package publisher

import "go.uber.org/fx"

var Module = fx.Module("publisher",
	fx.Provide(
		fx.Annotate(
			NewKafkaPublisher,
			fx.As(new(EventPublisher)),
		),
	),
)
