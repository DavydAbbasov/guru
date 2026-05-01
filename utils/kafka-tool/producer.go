package kafka

import (
	"context"

	"guru/utils/logger"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Producer struct {
	topic    string
	producer sarama.SyncProducer
	log      logger.Logger
}

func NewProducer(cfg *Config, log logger.Logger) (*Producer, error) {
	saramaCfg := sarama.NewConfig()
	saramaCfg.Version = sarama.V2_8_0_0
	saramaCfg.ClientID = cfg.ClientID

	saramaCfg.Producer.Return.Successes = true
	saramaCfg.Producer.Return.Errors = true
	saramaCfg.Producer.RequiredAcks = sarama.WaitForAll
	saramaCfg.Producer.Retry.Max = 5
	saramaCfg.Producer.Idempotent = true
	saramaCfg.Producer.Compression = sarama.CompressionSnappy
	saramaCfg.Net.MaxOpenRequests = 1 // required by Sarama when Idempotent=true

	producer, err := sarama.NewSyncProducer(cfg.Brokers, saramaCfg)
	if err != nil {
		if log != nil {
			log.Error("failed to create kafka producer",
				zap.String("clientId", cfg.ClientID),
				zap.Strings("brokers", cfg.Brokers),
				zap.Error(err),
			)
		}
		return nil, err
	}

	return &Producer{
		topic:    cfg.Topic,
		producer: producer,
		log:      log,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, key string, value []byte) (partition int32, offset int64, err error) {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	otel.GetTextMapPropagator().Inject(ctx, producerMessageCarrier{msg: msg})

	partition, offset, err = p.producer.SendMessage(msg)
	if err != nil && p.log != nil {
		p.log.Error("kafka send failed",
			zap.String("topic", p.topic),
			zap.String("key", key),
			zap.Error(err),
		)
	}
	return partition, offset, err
}

func (p *Producer) Topic() string {
	return p.topic
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
