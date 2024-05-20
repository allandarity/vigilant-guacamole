package letterboxd

import (
	"encoding/csv"
	"go-jellyfin-api/pkg/model"
	redisClient "go-jellyfin-api/pkg/redis"
	"golang.org/x/exp/rand"
	"os"
	"strconv"
	"time"
)

type Service interface {
	GetItemsFromWatchlistCSV() (model.Items, error)
	GetRandomNumberOfItemsFromWatchlist(int) []model.ItemsElement
	GetNumberOfKeysMatchingYear(noOfKeys int) ([]map[string][]model.ItemsElement, error)
}

type letterboxdService struct {
	rClient          redisClient.Client
	watchlistCSVPath string
	watchlistItems   []model.ItemsElement
}

func (l *letterboxdService) GetItemsFromWatchlistCSV() (model.Items, error) {
	file, err := os.Open(l.watchlistCSVPath)
	if err != nil {
		return model.Items{}, err
	}

	defer file.Close()

	reader := csv.NewReader(file)

	//skip the header line
	_, err = reader.Read()
	if err != nil {
		return model.Items{}, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return model.Items{}, err
	}
	var items model.Items

	for _, record := range records {
		name := record[1]
		year, _ := strconv.Atoi(record[2])
		item := &model.ItemsElement{
			Name:           name,
			ProductionYear: int16(year),
		}
		item.Name = item.NormaliseTitle()
		items.ItemElements = append(items.ItemElements, *item)
	}
	l.watchlistItems = items.ItemElements
	return items, nil
}

func (l *letterboxdService) GetRandomNumberOfItemsFromWatchlist(noOfItems int) []model.ItemsElement {
	var items model.Items
	rand.Seed(uint64(time.Now().UnixNano()))
	for i := 0; i < noOfItems; i++ {
		index := rand.Intn(len(l.watchlistItems))
		item := l.watchlistItems[index]
		if items.GetItemByName(item.Name).Name != item.Name {
			items.ItemElements = append(items.ItemElements, item)
		}
	}
	return items.ItemElements
}

func (l *letterboxdService) GetNumberOfKeysMatchingYear(noOfKeys int) ([]map[string][]model.ItemsElement, error) {
	output := make([]map[string][]model.ItemsElement, 0, noOfKeys)

	for len(output) < noOfKeys {
		watchlistItems := l.GetRandomNumberOfItemsFromWatchlist(10)

		for _, item := range watchlistItems {
			name := item.NormaliseTitle()
			keys, err := l.rClient.FindKeyByPartialTitle(name)
			if err != nil {
				return nil, err
			}

			extractedKeys := make([]model.ItemsElement, 0)
			for _, key := range keys {
				if key.ProductionYear == item.ProductionYear {
					extractedKeys = append(extractedKeys, key)
				}
			}

			if len(extractedKeys) > 0 {
				output = append(output, map[string][]model.ItemsElement{
					name: extractedKeys,
				})

				if len(output) >= noOfKeys {
					return output, nil
				}
			}
		}
	}

	return output, nil
}

func NewService(rClient redisClient.Client, watchlistCSVPath string) (Service, error) {
	return &letterboxdService{
		rClient:          rClient,
		watchlistCSVPath: watchlistCSVPath,
		watchlistItems:   nil,
	}, nil
}
