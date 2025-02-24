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

type AppConfig struct {
	DBPool              *pgxpool.Pool
	JellyfinConfig      config.JellyfinConfiguration
	JellyfinClient      jellyfinHttp.Client
	MovieFolderParentID string
}

type Services struct {
	Jellyfin       service.JellyfinService
	Movie          service.MovieService
	Watchlist      service.WatchlistService
	MovieWatchlist service.MovieWatchlistService
}

type Repositories struct {
	Movie          repository.MovieRepository
	Watchlist      repository.WatchlistRepository
	MovieWatchlist repository.MovieWatchlistRepository
}

func initializeConfig(ctx context.Context) (*AppConfig, error) {
	pool, err := initDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	jellyfinConfig, err := config.NewJellyfinConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create Jellyfin configuration: %w", err)
	}

	jellyfinClient, err := createJellyfinClient(jellyfinConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jellyfin client: %w", err)
	}

	movieFolderParentID, err := jellyfinClient.GetMovieFolderParentId()
	if err != nil {
		return nil, fmt.Errorf("failed to get movie folder parent ID: %w", err)
	}

	return &AppConfig{
		DBPool:              pool,
		JellyfinConfig:      jellyfinConfig,
		JellyfinClient:      jellyfinClient,
		MovieFolderParentID: movieFolderParentID,
	}, nil
}

func initializeRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Movie:          repository.NewMovieRepository(pool),
		Watchlist:      repository.NewWatchlistRepository(pool),
		MovieWatchlist: repository.NewMovieWatchlistRepository(pool),
	}
}

func initializeServices(config *AppConfig, repos *Repositories) *Services {
	return &Services{
		Jellyfin:  service.NewJellyfinService(config.JellyfinConfig, repos.Movie),
		Movie:     service.NewMovieService(repos.Movie),
		Watchlist: service.NewWatchlistService(repos.Watchlist),
		MovieWatchlist: service.NewMovieWatchlistService(
			service.NewMovieService(repos.Movie),
			service.NewWatchlistService(repos.Watchlist),
			repos.MovieWatchlist,
		),
	}
}

func syncMovieData(ctx context.Context, config *AppConfig, repos *Repositories) error {
	log.Println("Fetching movies from Jellyfin...")
	movies, err := config.JellyfinClient.GetAllMoviesRequest(config.MovieFolderParentID)
	if err != nil {
		return fmt.Errorf("failed to get movies: %w", err)
	}

	log.Println("Updating movie images...")
	moviesWithImages, err := config.JellyfinClient.PopulateMovieImageData(movies)
	if err != nil {
		return fmt.Errorf("failed to update movie images: %w", err)
	}

	log.Println("Populating movie database...")
	if err := repos.Movie.PopulateMovieDatabase(ctx, moviesWithImages); err != nil {
		return fmt.Errorf("failed to populate movie database: %w", err)
	}

	return nil
}

func syncWatchlistData(ctx context.Context, services *Services) error {
	log.Println("Loading watchlist from CSV...")
	if _, err := services.Watchlist.LoadWatchlistCSV(ctx); err != nil {
		return fmt.Errorf("failed to load watchlist CSV: %w", err)
	}

	log.Println("Creating movie-watchlist join table...")
	mwl, err := services.MovieWatchlist.PopulateDatabase(ctx)
	if err != nil {
		return fmt.Errorf("failed to populate movie-watchlist database: %w", err)
	}
	log.Printf("Movie watchlist populated: %v\n", mwl)

	return nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	config, err := initializeConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer config.DBPool.Close()

	repos := initializeRepositories(config.DBPool)
	services := initializeServices(config, repos)

	if err := syncMovieData(ctx, config, repos); err != nil {
		log.Fatal(err)
	}

	if err := syncWatchlistData(ctx, services); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting HTTP server...")
	if err := createHttpMux(
		services.Jellyfin,
		config.JellyfinConfig,
		config.JellyfinClient,
		services.MovieWatchlist,
	); err != nil {
		log.Fatal(err)
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

func createHttpMux(jService service.JellyfinService, jCfg config.JellyfinConfiguration,
	hClient jellyfinHttp.Client, mwlService service.MovieWatchlistService,
) error {
	cfg := jellyfinHttp.Config{
		JellyfinConfiguration: jCfg,
		JellyfinService:       jService,
		HttpClient:            hClient,
		MovieWatchlistService: mwlService,
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
