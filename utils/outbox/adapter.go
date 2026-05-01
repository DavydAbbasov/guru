package outbox

import (
	"context"

	kafka "guru/utils/kafka-tool"
)

type kafkaAdapter struct {
	producer *kafka.Producer
}

func NewKafkaPublisher(producer *kafka.Producer) Publisher {
	return &kafkaAdapter{producer: producer}
}

func (a *kafkaAdapter) Publish(ctx context.Context, key string, payload []byte, headers map[string]string) error {
	_, _, err := a.producer.SendMessage(ctx, key, payload, headers)
	return err
}
