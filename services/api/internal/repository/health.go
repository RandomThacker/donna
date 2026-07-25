package repository

import (
	"context"

	"github.com/RandomThacker/donna/services/api/internal/constant"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HealthRepository checks database reachability.
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	pool *pgxpool.Pool
}

// NewHealthRepository constructs a HealthRepository backed by pgxpool.
func NewHealthRepository(pool *pgxpool.Pool) HealthRepository {
	return &healthRepository{pool: pool}
}

func (r *healthRepository) Ping(ctx context.Context) error {
	var one int
	return r.pool.QueryRow(ctx, constant.SQLPing).Scan(&one)
}
