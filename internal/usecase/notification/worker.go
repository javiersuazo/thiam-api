package notification

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/evrone/go-clean-template/internal/entity/notification"
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

func (w *Worker) Start(ctx context.Context) error {
	eventsCh, err := w.bus.Subscribe(ctx, event.TopicTranslation)
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
	switch e.Type {
	case event.TypeTranslationCreated:
		w.handleTranslationCreated(ctx, e)
	default:
		w.log.Info("Unhandled event type: %s", e.Type)
	}
}

func (w *Worker) handleTranslationCreated(ctx context.Context, e eventbus.Event) {
	var payload event.TranslationCreatedPayload

	if err := json.Unmarshal(e.Payload, &payload); err != nil {
		w.log.Error(err, "Worker - handleTranslationCreated - json.Unmarshal")

		return
	}

	msg := &notification.InAppMessage{
		UserID: payload.UserID,
		Type:   "translation_completed",
		Title:  "Translation Complete",
		Body:   fmt.Sprintf("Your translation from %s to %s is ready", payload.Source, payload.Destination),
		Data: map[string]string{
			"translation_id": payload.TranslationID.String(),
		},
	}

	if err := w.service.SendInApp(ctx, msg); err != nil {
		w.log.Error(err, "Worker - handleTranslationCreated - w.service.SendInApp")
	}
}
