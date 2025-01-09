package main

import (
	"context"
	"errors"
	"fmt"
	jellyfinHttpClient "go-jellyfin-api/cmd/http"
	"go-jellyfin-api/cmd/jellyfin"
	"go-jellyfin-api/cmd/repository"
	"go-jellyfin-api/cmd/service"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TODO: clean this main func up
func main() {
	// TODO: is this timeout acceptable?
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pool, err := initDatabase()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	movieRepository := repository.NewMovieRepository(pool)

	jellyfinClient, err := jellyfin.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	jellyfinHttpClient, err := createJellyfinHttpClient(jellyfinClient)
	if err != nil {
		log.Fatal(err)
	}

	movieFolderParentId, err := jellyfinHttpClient.GetMovieFolderParentId()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("getting movies")
	allMovies, err := jellyfinHttpClient.GetAllMoviesRequest(movieFolderParentId)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("updating images")
	allMoviesImageUpdated, err := jellyfinHttpClient.PopualateMovieImageData(allMovies)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("populating database")
	if err := movieRepository.PopulateMovieDatabase(ctx, allMoviesImageUpdated); err != nil {
		log.Fatal(err)
	}

	fmt.Println("reading csv file")
	watchlistRepository := repository.NewWatchlistRepository(pool)
	watchlistService := service.NewWatchlistService(watchlistRepository)
	list, err := watchlistService.LoadWatchlistCSV()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list)

	fmt.Println("starting mux")
	httpErr := createHttpMux(jellyfinClient, movieRepository, jellyfinHttpClient)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
}

func createJellyfinHttpClient(jClient jellyfin.Client) (jellyfinHttpClient.Client, error) {
	jHttpClient, err := jellyfinHttpClient.NewClient(jClient)
	if err != nil {
		fmt.Println("Failed to create jellyfinHttpClient")
		return nil, err
	}

	authError := jHttpClient.AuthenticateByName()
	if authError != nil {
		fmt.Println("Failed to auth")
		return nil, authError
	}

	return jHttpClient, nil
}

func createHttpMux(jClient jellyfin.Client, db repository.MovieRepository, hClient jellyfinHttpClient.Client) error {
	cfg := jellyfinHttpClient.Config{
		JellyfinClient:  jClient,
		MovieRepository: db,
		HttpClient:      hClient,
	}
	rc := jellyfinHttpClient.NewController(cfg)

	httpServeError := http.ListenAndServe(":8080", rc.DefineMiddleware(rc.GetMux()))
	if httpServeError != nil {
		return httpServeError
	}
	return nil
}

func initDatabase() (*pgxpool.Pool, error) {
	databaseHost := os.Getenv("DATABASE_HOST")
	databasePort := os.Getenv("DATABASE_PORT")
	if databasePort == "" || databaseHost == "" {
		return nil, errors.New("DATABASE HOST or PORT not set")
	}

	databaseUrl := fmt.Sprintf(
		"postgres://user:password@%s:%s/movies?sslmode=disable",
		databaseHost,
		databasePort,
	)

	migration, err := migrate.New("file://db/migrations", databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("create migration: %w", err)
	}
	if err := migration.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("run migration: %w", err)
	}

	config, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}
