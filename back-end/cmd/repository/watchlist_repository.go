package repository

import "github.com/jackc/pgx/v5/pgxpool"

type WatchlistRepository interface{}

type watchlistRepository struct {
	pool *pgxpool.Pool
}

func NewWatchlistRepository(pool *pgxpool.Pool) WatchlistRepository {
	return &watchlistRepository{
		pool: pool,
	}
}
