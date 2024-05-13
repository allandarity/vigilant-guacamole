package tests

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCompareNamesFromCSVAndRedis(t *testing.T) {
	ctx := context.Background()

	c := setUpTestContainers(ctx)

	defer func() {
		if err := c.Terminate(ctx); err != nil {
			log.Fatalf("Could not stop redis: %s", err)
		}
	}()

	// Get the Redis container host and port
	redisHost, err := c.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis container host: %v", err)
	}
	redisPort, err := c.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get Redis container port: %v", err)
	}

	fmt.Println(redisHost, redisPort)

}

func setUpTestContainers(ctx context.Context) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Could not start redis: %s", err)
	}
	return redisC
}
