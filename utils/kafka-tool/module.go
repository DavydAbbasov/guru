package kafka

import (
	"context"

	"go.uber.org/fx"
)

var ProducerModule = fx.Module("kafka-producer",
	fx.Provide(NewProducer),
	fx.Invoke(registerProducerLifecycle),
)

// ConsumerModule provides a *Consumer but does not start it; the owning service binds the handler and Start/Stop.
var ConsumerModule = fx.Module("kafka-consumer",
	fx.Provide(NewConsumer),
)

func registerProducerLifecycle(lc fx.Lifecycle, p *Producer) {
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return p.Close()
		},
	})
}
