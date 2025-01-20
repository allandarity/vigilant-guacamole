package repository

import (
	"context"
	"fmt"
	"go-jellyfin-api/cmd/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieWatchlistRepository interface {
	InsertPairs(ctx context.Context, pairs []model.MovieWatchlistPair) error
}

type movieWatchlistRepository struct {
	pool *pgxpool.Pool
}

func NewMovieWatchlistRepository(pool *pgxpool.Pool) MovieWatchlistRepository {
	return &movieWatchlistRepository{
		pool: pool,
	}
}

func (m *movieWatchlistRepository) InsertPairs(ctx context.Context, pairs []model.MovieWatchlistPair) error {
	query := `
			INSERT INTO movie_watchlist (movie_id, watchlist_id, added_date)
			VALUES ($1, $2, $3)
			ON CONFLICT (movie_id, watchlist_id) DO NOTHING
	`
	batch := &pgx.Batch{}
	for _, pair := range pairs {
		batch.Queue(query, pair.MovieId, pair.WatchlistId, pair.AddedDate)
	}

	if batch.Len() > 0 {
		br := m.pool.SendBatch(ctx, batch)
		defer br.Close()
		if err := br.Close(); err != nil {
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	return nil
}
