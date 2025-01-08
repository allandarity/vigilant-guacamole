package http

import (
	"encoding/json"
	"fmt"
	"go-jellyfin-api/cmd/jellyfin"
	"go-jellyfin-api/cmd/repository"
	"net/http"
	"strconv"
)

type Controller interface {
	DefineRoutes()
	DefineMiddleware(next http.Handler) http.Handler
	GetMux() *http.ServeMux
}

type restController struct {
	mux             *http.ServeMux
	jellyfinClient  jellyfin.Client
	movieRepository repository.MovieRepository
	httpClient      Client
}

type Config struct {
	JellyfinClient  jellyfin.Client
	MovieRepository repository.MovieRepository
	HttpClient      Client
}

func NewController(cfg Config) Controller {
	c := &restController{
		mux:             http.NewServeMux(),
		jellyfinClient:  cfg.JellyfinClient,
		movieRepository: cfg.MovieRepository,
		httpClient:      cfg.HttpClient,
	}
	c.DefineRoutes()
	return c
}

func (c *restController) DefineRoutes() {
	c.mux.HandleFunc(
		"/movies/random",
		c.GetRandomMovies(3),
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

func (c restController) handleIncomingRequestForImage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		imageData, err := c.getItemImage(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", strconv.Itoa(len(imageData)))

		_, err = w.Write(imageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// TODO delete
func (c restController) getItemImage(itemId string) ([]byte, error) {
	getImageUrl := fmt.Sprintf("%s/Items/%s/Images/Primary?MaxWidth=%d&MaxHeight=%d",
		c.jellyfinClient.GetHost(), itemId, ImageMaxWidth, ImageMaxHeight)
	req, err := c.httpClient.GetRequest(getImageUrl)
	if err != nil {
		fmt.Errorf("Error creating request", "url", getImageUrl, "error", err)
		return []byte(""), err
	}
	resp, err := c.httpClient.MakeHttpClientRequest(req)
	if err != nil {
		fmt.Errorf("Error making HTTP request", "url", getImageUrl, "error", err)
		return []byte(""), err
	}
	return resp, nil
}

func (c restController) GetRandomMovies(count int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		movies, err := c.movieRepository.GetRandomMovies(ctx, count)
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
