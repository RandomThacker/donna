package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier is the subset of pgx pool/tx used by repositories.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// TxManager runs work inside a database transaction.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager constructs a TxManager.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// WithinTx begins a transaction, runs fn, and commits or rolls back.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := fn(ctx, tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
