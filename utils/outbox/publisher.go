package outbox

import "context"

type Publisher interface {
	Publish(ctx context.Context, key string, payload []byte, headers map[string]string) error
}
