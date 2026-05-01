package kafka

import (
	"context"
	"fmt"
	"time"

	"guru/utils/logger"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

const shutdownDrainTimeout = 30 * time.Second

type Consumer struct {
	group  sarama.ConsumerGroup
	topic  string
	ctx    context.Context
	cancel context.CancelFunc
	log    logger.Logger
	done   chan struct{}
}

func NewConsumer(cfg *Config, log logger.Logger) (*Consumer, error) {
	if cfg.GroupID == "" {
		return nil, fmt.Errorf("kafka consumer requires non-empty groupId")
	}

	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_8_0_0
	saramaCfg.ClientID = cfg.ClientID
	saramaCfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRange(),
	}
	saramaCfg.Consumer.Offsets.AutoCommit.Enable = false

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, saramaCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		group:  group,
		topic:  cfg.Topic,
		ctx:    ctx,
		cancel: cancel,
		log:    log,
		done:   make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(handler func(ctx context.Context, msg *sarama.ConsumerMessage) error) {
	go func() {
		defer close(c.done)
		for {
			if c.ctx.Err() != nil {
				return
			}
			h := &consumerHandler{
				handlerFunc: handler,
				log:         c.log,
			}
			if err := c.group.Consume(c.ctx, []string{c.topic}, h); err != nil {
				if c.log != nil {
					c.log.Error("kafka consume error",
						zap.String("topic", c.topic),
						zap.Error(err),
					)
				}
			}
		}
	}()
}

func (c *Consumer) Close() error {
	c.cancel()

	select {
	case <-c.done:
	case <-time.After(shutdownDrainTimeout):
		if c.log != nil {
			c.log.Warn("kafka consumer drain timed out, forcing shutdown",
				zap.String("topic", c.topic),
				zap.Duration("timeout", shutdownDrainTimeout),
			)
		}
	}

	return c.group.Close()
}

type consumerHandler struct {
	handlerFunc func(ctx context.Context, msg *sarama.ConsumerMessage) error
	log         logger.Logger
}

func (h *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		// recover so a handler panic doesn't kill the consume goroutine and orphan the group
		func() {
			defer func() {
				if r := recover(); r != nil {
					if h.log != nil {
						h.log.Error("panic in kafka handler",
							zap.String("topic", msg.Topic),
							zap.Int32("partition", msg.Partition),
							zap.Int64("offset", msg.Offset),
							zap.Any("panic", r),
						)
					}
				}
			}()

			ctx := otel.GetTextMapPropagator().Extract(session.Context(), consumerMessageCarrier{msg: msg})

			if err := h.handlerFunc(ctx, msg); err != nil {
				if h.log != nil {
					h.log.Error("kafka handler error",
						zap.String("topic", msg.Topic),
						zap.Int32("partition", msg.Partition),
						zap.Int64("offset", msg.Offset),
						zap.Error(err),
					)
				}
				// TODO: route to dead-letter topic; mark+commit below avoids a poison-message redelivery loop
			}
			session.MarkMessage(msg, "")
			session.Commit()
		}()
	}
	return nil
}
