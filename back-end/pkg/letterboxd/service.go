package letterboxd

import (
	"encoding/csv"
	"go-jellyfin-api/pkg/model"
	redisClient "go-jellyfin-api/pkg/redis"
	"os"
	"strconv"
)

type Service interface {
	GetItemsFromWatchlistCSV() (model.Items, error)
}

type letterboxdService struct {
	lHttpClient      Client
	rClient          redisClient.Client
	watchlistCSVPath string
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
		name := l.rClient.NormaliseTitle(record[1])
		year, _ := strconv.Atoi(record[2])
		item := model.ItemsElement{
			Name:           name,
			ProductionYear: int16(year),
		}

		items.ItemElements = append(items.ItemElements, item)
	}

	return items, nil
}

func NewService(rClient redisClient.Client, watchlistCSVPath string) (Service, error) {
	return &letterboxdService{
		lHttpClient:      letterboxdHttpClient{},
		rClient:          rClient,
		watchlistCSVPath: watchlistCSVPath,
	}, nil
}
