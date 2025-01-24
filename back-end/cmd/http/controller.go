package http

import (
	"encoding/json"
	"fmt"
	"go-jellyfin-api/cmd/config"
	"go-jellyfin-api/cmd/service"
	"net/http"
)

type Controller interface {
	DefineRoutes()
	DefineMiddleware(next http.Handler) http.Handler
	GetMux() *http.ServeMux
}

type restController struct {
	mux                   *http.ServeMux
	jellyfinService       service.JellyfinService
	httpClient            Client
	jellyfinConfiguration config.JellyfinConfiguration
	movieWatchlistService service.MovieWatchlistService
}

type Config struct {
	JellyfinConfiguration config.JellyfinConfiguration
	JellyfinService       service.JellyfinService
	HttpClient            Client
	MovieWatchlistService service.MovieWatchlistService
}

func NewController(cfg Config) Controller {
	c := &restController{
		mux:                   http.NewServeMux(),
		jellyfinService:       cfg.JellyfinService,
		httpClient:            cfg.HttpClient,
		jellyfinConfiguration: cfg.JellyfinConfiguration,
		movieWatchlistService: cfg.MovieWatchlistService,
	}
	c.DefineRoutes()
	return c
}

func (c restController) DefineRoutes() {
	c.mux.HandleFunc(
		"/movies/random",
		c.GetRandomMovies(3),
	)
	c.mux.HandleFunc(
		"/movies/watchlist/random",
		c.GetRandomMoviesFromWatchlist(3),
	)
}

func (c restController) DefineMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		next.ServeHTTP(w, r)
	})
}

func (c restController) GetMux() *http.ServeMux {
	return c.mux
}

func (c restController) GetRandomMovies(count int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		ctx := r.Context()
		movies, err := c.jellyfinService.GetRandomMovies(ctx, count)
		if err != nil {
			fmt.Printf("Error getting random movies", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		jsonBody, err := json.Marshal(movies)
		if err != nil {
			fmt.Printf("Error marshalling movies to JSON", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonBody)
		if err != nil {
			fmt.Printf("Error writing response body", "error", err)
		}
	}
}

func (c restController) GetRandomMoviesFromWatchlist(count int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
		ctx := r.Context()
		movies, err := c.movieWatchlistService.GetRandomMovieWatchlist(ctx, count)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		jsonBody, err := json.Marshal(movies)
		if err != nil {
			fmt.Printf("Error marshalling movies to JSON", "error", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(jsonBody)
		if err != nil {
			fmt.Printf("Error writing response body", "error", err)
		}
	}
}
