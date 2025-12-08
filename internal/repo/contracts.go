// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/entity/event"
	"github.com/google/uuid"
)

//go:generate mockgen -source=contracts.go -destination=../usecase/mocks_repo_test.go -package=usecase_test

type (
	// TranslationRepo -.
	TranslationRepo interface {
		Store(context.Context, entity.Translation) error
		GetHistory(context.Context) ([]entity.Translation, error)
	}

	// TranslationWebAPI -.
	TranslationWebAPI interface {
		Translate(entity.Translation) (entity.Translation, error)
	}

	// OutboxRepo handles outbox event persistence.
	OutboxRepo interface {
		Store(ctx context.Context, events []event.OutboxEvent) error
		FetchUnpublished(ctx context.Context, limit int) ([]event.OutboxEvent, error)
		MarkPublished(ctx context.Context, id uuid.UUID) error
		MarkFailed(ctx context.Context, id uuid.UUID, err error) error
	}
)
