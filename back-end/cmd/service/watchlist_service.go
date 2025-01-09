package service

import (
	"encoding/csv"
	"fmt"
	"go-jellyfin-api/cmd/model"
	"go-jellyfin-api/cmd/repository"
	"os"
)

const (
	RESOURCE_LOCATION  = "../../resources/"
	WATCHLIST_FILENAME = "watchlist.csv"
)

type WatchlistService interface {
	LoadWatchlistCSV() (model.Watchlist, error)
}

type watchlistService struct {
	repository repository.WatchlistRepository
}

func NewWatchlistService(repository repository.WatchlistRepository) WatchlistService {
	return &watchlistService{
		repository: repository,
	}
}

func (w *watchlistService) LoadWatchlistCSV() (model.Watchlist, error) {
	path := RESOURCE_LOCATION + WATCHLIST_FILENAME
	result, err := readCsvFile(path)
	if err != nil {
		return model.Watchlist{}, err
	}
	return result, nil
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
		var item model.WatchlistItem
		item.DateAdded = record[0]
		item.Title = record[1]
		item.DateReleased = record[2]
		watchlistItems = append(watchlistItems, item)
	}
	watchlist.Items = watchlistItems
	return watchlist
}
