package main

import (
	"context"
	"fmt"
	jellyfinHttpClient "go-jellyfin-api/pkg/http"
	"go-jellyfin-api/pkg/jellyfin"
	redisClient "go-jellyfin-api/pkg/redis"
	"net/http"
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
	fmt.Println(allMovies)
	addMoviesErr := rClient.AddItems(allMovies)

	if addMoviesErr != nil {
		fmt.Println(addMoviesErr)
	}

	httpErr := createHttpMux(jClient, rClient, hClient)
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

func createHttpMux(jClient jellyfin.Client, rClient redisClient.Client, hClient jellyfinHttpClient.Client) error {

	rc := jellyfinHttpClient.NewController(jClient, rClient, hClient)
	rc.DefineRoutes(rc.GetMux())

	httpServeError := http.ListenAndServe(":8080", rc.DefineMiddleware(rc.GetMux()))
	if httpServeError != nil {
		return httpServeError
	}
	return nil
}
