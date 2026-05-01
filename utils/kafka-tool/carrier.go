package kafka

import "github.com/IBM/sarama"

type producerMessageCarrier struct {
	msg *sarama.ProducerMessage
}

func (c producerMessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c producerMessageCarrier) Set(key, value string) {
	for i, h := range c.msg.Headers {
		if string(h.Key) == key {
			c.msg.Headers[i].Value = []byte(value)
			return
		}
	}
	c.msg.Headers = append(c.msg.Headers, sarama.RecordHeader{
		Key:   []byte(key),
		Value: []byte(value),
	})
}

func (c producerMessageCarrier) Keys() []string {
	keys := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		keys[i] = string(h.Key)
	}
	return keys
}

type consumerMessageCarrier struct {
	msg *sarama.ConsumerMessage
}

func (c consumerMessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c consumerMessageCarrier) Set(key, value string) {}

func (c consumerMessageCarrier) Keys() []string {
	keys := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		keys[i] = string(h.Key)
	}
	return keys
}
