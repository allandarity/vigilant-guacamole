package letterboxd

import (
	"context"
	"go-jellyfin-api/pkg/model"
	"go-jellyfin-api/pkg/redis"
	"os"
	"reflect"
	"testing"
)

var testData = []byte(`
    Date,Name,Year,Letterboxd URI
    2020-04-26,Lady Bird,2017,https://boxd.it/dGNE
    2020-04-26,Columbus,2017,https://boxd.it/eCuA
    2020-04-26,Let the Corpses Tan,2017,https://boxd.it/eWSs
`)

func TestReadCSVFile(t *testing.T) {

	tempFile, err := os.CreateTemp("", "test.csv")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	t.Logf("Temporary file path: %s", tempFile.Name())

	if _, err := tempFile.Write(testData); err != nil {
		t.Fatalf("Failed to write test data to temporary file: %v", err)
	}

	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}
	ctx := context.Background()

	service := &letterboxdService{
		watchlistCSVPath: tempFile.Name(),
		rClient:          redis.NewClient(ctx),
	}

	result, err := service.GetItemsFromWatchlistCSV()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result.IsEmpty() {
		t.Error("Resulting output is empty")
	}

	expectedType := reflect.TypeOf(model.Items{})
	resultingType := reflect.TypeOf(result)
	if resultingType != expectedType {
		t.Errorf("Expected type %s but got %s", expectedType, resultingType)
	}

	expectedItems := []model.ItemsElement{
		{Name: "lady_bird", ProductionYear: 2017},
		{Name: "columbus", ProductionYear: 2017},
		{Name: "let_the_corpses_tan", ProductionYear: 2017},
	}

	if len(result.ItemElements) != len(expectedItems) {
		t.Errorf("Expected %d items but got %d", len(expectedItems), len(result.ItemElements))
	}

	for i, item := range result.ItemElements {
		expected := expectedItems[i]
		if item.Name != expected.Name || item.ProductionYear != expected.ProductionYear {
			t.Errorf("Item %d does not match expected values", i)
		}
	}
}

/*
package letterboxd

import (
    "testing"

    redisClient "go-jellyfin-api/pkg/redis"
    "go-jellyfin-api/pkg/model"
)

func TestCompareNamesFromCSVAndRedis(t *testing.T) {
    // Set up test data
    csvNames := []string{"movie1", "movie2", "movie3"}
    redisItems := []model.ItemsElement{
        {Name: "movie1"},
        {Name: "movie4"},
        {Name: "movie5"},
    }

    // Create a mock Redis client
    mockRedisClient := &mockRedisClient{
        items: redisItems,
    }

    // Create a mock Letterboxd service
    mockLetterboxdService := &mockLetterboxdService{
        csvNames: csvNames,
    }

    // Test the comparison
    unmatchedNames := findUnmatchedNames(mockLetterboxdService.csvNames, mockRedisClient.getAllMovieNames())

    // Assert the expected result
    expectedUnmatchedNames := []string{"movie2", "movie3", "movie4", "movie5"}
    if len(unmatchedNames) != len(expectedUnmatchedNames) {
        t.Errorf("Expected %d unmatched names, but got %d", len(expectedUnmatchedNames), len(unmatchedNames))
    }

    for i, name := range unmatchedNames {
        if name != expectedUnmatchedNames[i] {
            t.Errorf("Unmatched name mismatch at index %d: expected %s, but got %s", i, expectedUnmatchedNames[i], name)
        }
    }
}

// Mock implementations for testing purposes
type mockRedisClient struct {
    items []model.ItemsElement
}

func (m *mockRedisClient) getAllMovieNames() []string {
    var names []string
    for _, item := range m.items {
        names = append(names, item.Name)
    }
    return names
}

type mockLetterboxdService struct {
    csvNames []string
}

func (m *mockLetterboxdService) ReadCSVFile() ([]string, error) {
    return m.csvNames, nil
}
*/
