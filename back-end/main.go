package main

import (
	"context"
	"fmt"
	jellyfinHttpClient "go-jellyfin-api/pkg/http"
	"go-jellyfin-api/pkg/jellyfin"
	"go-jellyfin-api/pkg/letterboxd"
	redisClient "go-jellyfin-api/pkg/redis"
	"net/http"
	"os"
)

func main() {

	jClient, err := jellyfin.NewClient()
	if err != nil {
		panic(err)
	}

	hClient, err := createJellyfinHttpClient(jClient)

	if err != nil {
		panic(err)
	}

	rClient := redisClient.NewClient(context.Background())

	movieFolderParentId, movieParentFolderErr := hClient.GetMovieFolderParentId()
	if movieParentFolderErr != nil {
		fmt.Println(movieParentFolderErr)
	}

	allMovies, allMoviesErr := hClient.GetAllMoviesRequest(movieFolderParentId)

	if allMoviesErr != nil {
		fmt.Println(allMoviesErr)
		panic(allMoviesErr)
	}
	addMoviesErr := rClient.AddItems(allMovies)

	if addMoviesErr != nil {
		fmt.Println(addMoviesErr)
	}

	csvPath := os.Getenv("WATCHLIST_CSV_PATH")
	if csvPath == "" {
		panic("WATCHLIST_CSV_PATH not set")
	}
	lService, _ := letterboxd.NewService(rClient, csvPath)

	items, err := rClient.GetItemsByKeyword("batman")
	if err != nil {
		fmt.Println("FAILED TO FIND BATMAN")
	}
	fmt.Println(items)
	httpErr := createHttpMux(jClient, rClient, hClient, lService)
	if httpErr != nil {
		panic(httpErr)
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

func createHttpMux(jClient jellyfin.Client, rClient redisClient.Client, hClient jellyfinHttpClient.Client, lService letterboxd.Service) error {

	rc := jellyfinHttpClient.NewController(jClient, rClient, hClient, lService)
	rc.DefineRoutes(rc.GetMux())

	httpServeError := http.ListenAndServe(":8080", rc.DefineMiddleware(rc.GetMux()))
	if httpServeError != nil {
		return httpServeError
	}
	return nil
}
