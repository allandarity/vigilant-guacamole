package letterboxd

import (
	"encoding/csv"
	"fmt"
	redisClient "go-jellyfin-api/pkg/redis"
	"os"
)

const (
	WATCHLIST_CSV = "/app/resources/watchlist.csv"
)

type Service interface {
	PopulateHttpClient()
	RunTest() string
	ReadCSVFile() (string, error)
}

type letterboxdService struct {
	lHttpClient Client
	rClient     redisClient.Client
}

func (l *letterboxdService) ReadCSVFile() (string, error) {
	file, err := os.Open(WATCHLIST_CSV)
	if err != nil {
		return "", err
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	for _, record := range records {
		fmt.Println(record[1])
	}

	return "yes", nil
}

func (l *letterboxdService) PopulateHttpClient() {
	service, _ := NewClient()
	fmt.Println(service.TestClient("OUTPUT"))
	l.lHttpClient = service
}

func (l letterboxdService) RunTest() string {
	fmt.Println("Running Test")
	return l.lHttpClient.TestClient("YES")
}

func NewService(rClient redisClient.Client) (Service, error) {
	return &letterboxdService{
		lHttpClient: letterboxdHttpClient{},
		rClient:     rClient,
	}, nil
}
