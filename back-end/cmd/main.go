package main

import (
	"context"
	"errors"
	"fmt"
	"go-jellyfin-api/cmd/config"
	jellyfinHttp "go-jellyfin-api/cmd/http"
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

	jellyfinConfiguration, err := config.NewJellyfinConfiguration()
	if err != nil {
		log.Fatal(err)
	}
	jellyfinService := service.NewJellyfinService(jellyfinConfiguration, movieRepository)
	jellyfinHttpClient, err := createJellyfinClient(jellyfinConfiguration)
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
	allMoviesImageUpdated, err := jellyfinHttpClient.PopulateMovieImageData(allMovies)
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
	_, err = watchlistService.LoadWatchlistCSV(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("creating join table")
	movieService := service.NewMovieService(movieRepository)
	movieWatchlistRepository := repository.NewMovieWatchlistRepository(pool)
	movieWatchlistService := service.NewMovieWatchlistService(movieService, watchlistService, movieWatchlistRepository)
	mwl, err := movieWatchlistService.PopulateDatabase(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mwl)

	fmt.Println("starting mux")
	httpErr := createHttpMux(jellyfinService, jellyfinConfiguration, jellyfinHttpClient)
	if httpErr != nil {
		log.Fatal(httpErr)
	}
}

func createJellyfinClient(cfg config.JellyfinConfiguration) (jellyfinHttp.Client, error) {
	jHttpClient, err := jellyfinHttp.NewClient(cfg)
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

func createHttpMux(jService service.JellyfinService, jCfg config.JellyfinConfiguration, hClient jellyfinHttp.Client) error {
	cfg := jellyfinHttp.Config{
		JellyfinConfiguration: jCfg,
		JellyfinService:       jService,
		HttpClient:            hClient,
	}
	rc := jellyfinHttp.NewController(cfg)

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
	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("run migration: %w", err)
	}

	pxgPoolConfig, err := pgxpool.ParseConfig(databaseUrl)
	if err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	pxgPoolConfig.MaxConns = 25
	pxgPoolConfig.MinConns = 5

	pool, err := pgxpool.NewWithConfig(context.Background(), pxgPoolConfig)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	return pool, nil
}
