package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/postgres"
	"github.com/google/uuid"
)

const _defaultEntityCap = 64

// TranslationRepo -.
type TranslationRepo struct {
	*postgres.Postgres
}

// New -.
func New(pg *postgres.Postgres) *TranslationRepo {
	return &TranslationRepo{pg}
}

// GetHistory -.
func (r *TranslationRepo) GetHistory(ctx context.Context) ([]entity.Translation, error) {
	sql, _, err := r.Builder.
		Select("id", "source", "destination", "original", "translation", "created_at", "updated_at").
		From("history").
		OrderBy("created_at DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Builder: %w", err)
	}

	rows, err := r.Pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Pool.Query: %w", err)
	}
	defer rows.Close()

	entities := make([]entity.Translation, 0, _defaultEntityCap)

	for rows.Next() {
		e := entity.Translation{}

		err = rows.Scan(&e.ID, &e.Source, &e.Destination, &e.Original, &e.Translation, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("TranslationRepo - GetHistory - rows.Scan: %w", err)
		}

		entities = append(entities, e)
	}

	return entities, nil
}

// Store -.
func (r *TranslationRepo) Store(ctx context.Context, t *entity.Translation) error {
	now := time.Now().UTC()

	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	t.CreatedAt = now
	t.UpdatedAt = now

	sql, args, err := r.Builder.
		Insert("history").
		Columns("id", "source", "destination", "original", "translation", "created_at", "updated_at").
		Values(t.ID, t.Source, t.Destination, t.Original, t.Translation, t.CreatedAt, t.UpdatedAt).
		ToSql()
	if err != nil {
		return fmt.Errorf("TranslationRepo - Store - r.Builder: %w", err)
	}

	_, err = r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("TranslationRepo - Store - r.Pool.Exec: %w", err)
	}

	return nil
}
