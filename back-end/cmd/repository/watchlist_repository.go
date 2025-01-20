package repository

import (
	"context"
	"fmt"
	"go-jellyfin-api/cmd/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WatchlistRepository interface {
	PopulateDatabase(ctx context.Context, items model.Watchlist) error
	GetAllWatchlist(ctx context.Context) ([]model.WatchlistItem, error)
}

type watchlistRepository struct {
	pool *pgxpool.Pool
}

func NewWatchlistRepository(pool *pgxpool.Pool) WatchlistRepository {
	return &watchlistRepository{
		pool: pool,
	}
}

func (w *watchlistRepository) PopulateDatabase(ctx context.Context, items model.Watchlist) error {
	batch := &pgx.Batch{}
	for _, item := range items.WatchlistItems {
		batch.Queue(
			`
      INSERT INTO watchlist(title, production_year, added_date)
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

func (w *watchlistRepository) GetAllWatchlist(ctx context.Context) ([]model.WatchlistItem, error) {
	query := `
		SELECT id, title, production_year, added_date from watchlist
	`
	rows, err := w.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var watchlist []model.WatchlistItem
	for rows.Next() {
		var item model.WatchlistItem
		if err := rows.Scan(&item.Id, &item.Title, &item.DateReleased, &item.DateAdded); err != nil {
			return nil, err
		}
		watchlist = append(watchlist, item)
	}
	return watchlist, nil
}
