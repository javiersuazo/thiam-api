package eventbus

import "context"

type Event struct {
	ID      string
	Type    string
	Payload []byte
}

type Subscriber interface {
	Subscribe(ctx context.Context, topic string) (<-chan Event, error)
	Close() error
}
