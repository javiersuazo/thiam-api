package eventbus

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/pkg/logger"
)

const (
	defaultPollInterval = time.Second
	defaultBatchSize    = 100
	defaultMaxRetries   = 5
)

type Worker struct {
	outboxRepo   repo.OutboxRepo
	publisher    Publisher
	logger       logger.Interface
	pollInterval time.Duration
	batchSize    int
	maxRetries   int
	stop         chan struct{}
	done         chan struct{}
}

type WorkerOption func(*Worker)

func WithPollInterval(d time.Duration) WorkerOption {
	return func(w *Worker) {
		w.pollInterval = d
	}
}

func WithBatchSize(size int) WorkerOption {
	return func(w *Worker) {
		w.batchSize = size
	}
}

func WithMaxRetries(maxRetries int) WorkerOption {
	return func(w *Worker) {
		w.maxRetries = maxRetries
	}
}

func NewWorker(outboxRepo repo.OutboxRepo, publisher Publisher, l logger.Interface, opts ...WorkerOption) *Worker {
	w := &Worker{
		outboxRepo:   outboxRepo,
		publisher:    publisher,
		logger:       l,
		pollInterval: defaultPollInterval,
		batchSize:    defaultBatchSize,
		maxRetries:   defaultMaxRetries,
		stop:         make(chan struct{}),
		done:         make(chan struct{}),
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

func (w *Worker) Start(ctx context.Context) {
	go w.run(ctx)
	w.logger.Info("eventbus worker - started")
}

func (w *Worker) Stop() {
	close(w.stop)
	<-w.done
	w.logger.Info("eventbus worker - stopped")
}

func (w *Worker) run(ctx context.Context) {
	defer close(w.done)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stop:
			return
		case <-ticker.C:
			if err := w.processOutbox(ctx); err != nil {
				w.logger.Error(err, "eventbus worker - process outbox")
			}
		}
	}
}

func (w *Worker) processOutbox(ctx context.Context) error {
	events, err := w.outboxRepo.FetchUnpublished(ctx, w.batchSize)
	if err != nil {
		return fmt.Errorf("fetch unpublished: %w", err)
	}

	for _, e := range events {
		if e.RetryCount >= w.maxRetries {
			w.logger.Warn(fmt.Sprintf("eventbus worker - max retries exceeded for event %s", e.ID))

			continue
		}

		if err := w.publisher.Publish(ctx, e); err != nil {
			w.logger.Error(err, fmt.Sprintf("eventbus worker - publish event %s", e.ID))

			if markErr := w.outboxRepo.MarkFailed(ctx, e.ID, err); markErr != nil {
				w.logger.Error(markErr, "eventbus worker - mark failed")
			}

			continue
		}

		if err := w.outboxRepo.MarkPublished(ctx, e.ID); err != nil {
			w.logger.Error(err, fmt.Sprintf("eventbus worker - mark published %s", e.ID))
		}
	}

	return nil
}
