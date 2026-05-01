package kafka

import (
	"context"

	"guru/backend/notification/internal/config"
	"guru/utils/logger"
	kafkatool "guru/utils/kafka-tool"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var Module = fx.Module("kafka-notification-consumer",
	kafkatool.ConsumerModule,
	fx.Provide(
		provideKafkaConfig,
		NewConsumerHandler,
	),
	fx.Invoke(startConsumer),
)

func provideKafkaConfig(cfg *config.KafkaConfig) *kafkatool.Config {
	return &kafkatool.Config{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	}
}

func startConsumer(
	lc fx.Lifecycle,
	consumer *kafkatool.Consumer,
	handler *ConsumerHandler,
	cfg *config.KafkaConfig,
	log logger.Logger,
) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Info("starting kafka consumer",
				zap.String("topic", cfg.Topic),
				zap.String("groupId", cfg.GroupID),
			)
			consumer.Start(handler.Handle)
			return nil
		},
		OnStop: func(_ context.Context) error {
			log.Info("stopping kafka consumer (graceful drain)")
			return consumer.Close()
		},
	})
}
