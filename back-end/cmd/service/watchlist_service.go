package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"go-jellyfin-api/cmd/model"
	"go-jellyfin-api/cmd/repository"
	"log"
	"os"
	"time"
)

const (
	RESOURCE_LOCATION  = "/app/resources/"
	WATCHLIST_FILENAME = "watchlist.csv"
)

type WatchlistService interface {
	LoadWatchlistCSV(ctx context.Context) (model.Watchlist, error)
	GetAllWatchlist(ctx context.Context) ([]model.WatchlistItem, error)
}

type watchlistService struct {
	repository repository.WatchlistRepository
}

func NewWatchlistService(repository repository.WatchlistRepository) WatchlistService {
	return &watchlistService{
		repository: repository,
	}
}

func (w *watchlistService) LoadWatchlistCSV(ctx context.Context) (model.Watchlist, error) {
	path := RESOURCE_LOCATION + WATCHLIST_FILENAME
	result, err := readCsvFile(path)
	if err != nil {
		return model.Watchlist{}, err
	}
	err = w.repository.PopulateDatabase(ctx, result)
	if err != nil {
		fmt.Println(err)
		return model.Watchlist{}, err
	}
	return result, nil
}

func (w *watchlistService) GetAllWatchlist(ctx context.Context) ([]model.WatchlistItem, error) {
	list, err := w.repository.GetAllWatchlist(ctx)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func readCsvFile(filePath string) (model.Watchlist, error) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("failed to open file at " + filePath)
		return model.Watchlist{}, nil
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("failed to read records")
		return model.Watchlist{}, nil
	}
	return mapToWatchlist(records), nil
}

func mapToWatchlist(records [][]string) model.Watchlist {
	var watchlist model.Watchlist
	var watchlistItems []model.WatchlistItem
	for _, record := range records {
		if record[0] == "Date" {
			// skip header
			continue
		}
		var item model.WatchlistItem
		item.DateAdded = parseTime("2006-01-02", record[0])
		item.Title = record[1]
		item.DateReleased = parseTime("2006", record[2])
		watchlistItems = append(watchlistItems, item)
	}
	watchlist.WatchlistItems = watchlistItems
	return watchlist
}

func parseTime(layout string, timeString string) time.Time {
	if timeString == "" {
		return time.Time{}
	}
	parsedTime, err := time.Parse(layout, timeString)
	if err != nil {
		log.Fatal(err)
		return time.Time{}
	}
	return parsedTime
}
