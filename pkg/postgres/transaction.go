package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX represents a database connection that can execute queries.
// Both *pgxpool.Pool and pgx.Tx implement this interface.
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxManager handles database transactions.
type TxManager struct {
	pool *Postgres
}

// NewTxManager creates a new transaction manager.
func NewTxManager(pg *Postgres) *TxManager {
	return &TxManager{pool: pg}
}

// WithTransaction executes the given function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (m *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := m.pool.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx) //nolint:errcheck // rollback error is less important than the original error
		}
	}()

	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}
