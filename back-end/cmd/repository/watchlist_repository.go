package repository

import (
	"context"
	"fmt"
	"go-jellyfin-api/cmd/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WatchlistRepository interface {
	PopulateDatabse(ctx context.Context, items model.Watchlist) error
}

type watchlistRepository struct {
	pool *pgxpool.Pool
}

func NewWatchlistRepository(pool *pgxpool.Pool) WatchlistRepository {
	return &watchlistRepository{
		pool: pool,
	}
}

func (w *watchlistRepository) PopulateDatabse(ctx context.Context, items model.Watchlist) error {
	batch := &pgx.Batch{}
	for _, item := range items.WatchlistItems {
		batch.Queue(
			`
      INSERT INTO watchlist(title, production_year, date_added)
      VALUES ($1, $2, $3)
      `,
			item.Title,
			item.DateReleased,
			item.DateAdded,
		)
	}

	if batch.Len() > 0 {
		br := w.pool.SendBatch(ctx, batch)
		defer br.Close()
		if err := br.Close(); err != nil {
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}
	return nil
}
