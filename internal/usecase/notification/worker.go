package notification

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/pkg/eventbus"
	"github.com/evrone/go-clean-template/pkg/logger"
)

type Worker struct {
	service *Service
	bus     eventbus.Subscriber
	log     logger.Interface
}

func NewWorker(service *Service, bus eventbus.Subscriber, log logger.Interface) *Worker {
	return &Worker{
		service: service,
		bus:     bus,
		log:     log,
	}
}

func (w *Worker) Start(ctx context.Context, topic string) error {
	eventsCh, err := w.bus.Subscribe(ctx, topic)
	if err != nil {
		return fmt.Errorf("Worker - Start - w.bus.Subscribe: %w", err)
	}

	go w.processEvents(ctx, eventsCh)

	return nil
}

func (w *Worker) processEvents(ctx context.Context, events <-chan eventbus.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case e, ok := <-events:
			if !ok {
				return
			}

			w.handleEvent(ctx, e)
		}
	}
}

func (w *Worker) handleEvent(ctx context.Context, e eventbus.Event) {
	w.log.Info("Received event type: %s", e.Type)

	// Event handlers will be implemented based on domain event types:
	// - user.created: Send welcome notification
	// - order.completed: Send order confirmation
	// - payment.failed: Send payment failure alert
	// For now, events are logged but not processed.
	_ = ctx
}
