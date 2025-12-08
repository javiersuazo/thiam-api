package eventbus

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity/event"
)

type Publisher interface {
	Publish(ctx context.Context, e *event.OutboxEvent) error
	Close() error
}
