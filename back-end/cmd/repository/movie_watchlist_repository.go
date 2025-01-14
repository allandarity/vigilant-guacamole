package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieWatchlistRepository interface {
	PopulateDatabase(ctx context.Context) error
}

type movieWatchlistRepository struct {
	pool *pgxpool.Pool
}

func NewMovieWatchlistRepository(pool *pgxpool.Pool) MovieWatchlistRepository {
	return &movieWatchlistRepository{
		pool: pool,
	}
}

func (m *movieWatchlistRepository) PopulateDatabase(ctx context.Context) error {
	batch := &pgx.Batch{}

	return nil
}
