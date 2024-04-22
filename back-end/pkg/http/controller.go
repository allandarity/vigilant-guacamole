package http

import (
	"encoding/json"
	"fmt"
	"go-jellyfin-api/pkg/jellyfin"
	redisClient "go-jellyfin-api/pkg/redis"
	"net/http"
	"strconv"
)

type Controller interface {
	DefineRoutes(mux *http.ServeMux)
	DefineMiddleware(next http.Handler) http.Handler
	GetMux() *http.ServeMux
	handleIncomingRequestForImage(w http.ResponseWriter, r *http.Request)
	getItemImage(itemId string) ([]byte, error)
}

type restController struct {
	mux     *http.ServeMux
	jClient jellyfin.Client
	rClient redisClient.Client
	hClient Client
}

func NewController(jClient jellyfin.Client, rClient redisClient.Client, hClient Client) Controller {
	mux := http.NewServeMux()
	return restController{
		mux:     mux,
		jClient: jClient,
		rClient: rClient,
		hClient: hClient,
	}
}

func (c restController) DefineRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleIncomingGetRequest(c.rClient, w)
	})
	mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
		c.handleIncomingRequestForImage(w, r)
	})
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

func (c restController) handleIncomingRequestForImage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
	imageData, err := c.getItemImage(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(imageData)))
	_, imageWriteErr := w.Write(imageData)
	if imageWriteErr != nil {
		http.Error(w, imageWriteErr.Error(), http.StatusInternalServerError)
		return
	}
}

func (c restController) getItemImage(itemId string) ([]byte, error) {
	getImageUrl := fmt.Sprintf("%s/Items/%s/Images/Primary?MaxWidth=400&MaxHeight=400", c.jClient.GetHost(), itemId)
	req, err := c.hClient.GetRequest(getImageUrl)
	if err != nil {
		fmt.Println(err)
		return []byte(""), err
	}
	resp, err := c.hClient.MakeHttpClientRequest(req)
	if err != nil {
		fmt.Println(err)
		return []byte(""), err
	}
	return resp, nil
}

func handleIncomingGetRequest(redisClient redisClient.Client, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	items, err := redisClient.GetRandomNumberOfItems(3)
	if err != nil {
		fmt.Println(err)
	}
	jsonBody, err := json.Marshal(items)
	if err != nil {
		fmt.Println(err)
	}
	w.Write(jsonBody)
}
