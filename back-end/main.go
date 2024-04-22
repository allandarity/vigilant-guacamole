package main

import (
	"context"
	"fmt"
	jellyfinHttpClient "go-jellyfin-api/pkg/http"
	"go-jellyfin-api/pkg/jellyfin"
	redisClient "go-jellyfin-api/pkg/redis"
	"net/http"
)

func oldMain() {
	//mediaBrowser, err := buildMediaBrowser()
	//if err != nil {
	//	panic(err)
	//}
	//authenticationRequest, err := buildAuthenticationRequest()
	//if err != nil {
	//	panic(err)
	//}
	//authRequest, err := makeAuthenticationRequest(mediaBrowser, authenticationRequest)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//authResponse, err := handleMakeAuthenticationRequest(authRequest)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//getMovieParentIdRequestUrl := fmt.Sprintf("%s/Users/%s/Items", mediaBrowser.host, authResponse.User.Id)
	//parentFolderRequest, err := makeGetRequest(getMovieParentIdRequestUrl, mediaBrowser)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//parentFolderResponse, err := handleGetMovieParentIdRequest(parentFolderRequest)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//fmt.Println(parentFolderResponse.ItemElements)
	//parentId, err := getMoviesParentId(parentFolderResponse)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//
	//getAllMoviesRequestUrl := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s", mediaBrowser.host, authResponse.User.Id, parentId)
	//allMovieRequest, err := makeGetRequest(getAllMoviesRequestUrl, mediaBrowser)
	//if err != nil {
	//	fmt.Println(err)
	//}
	//movieItems, handleErr := handleGetMovieParentIdRequest(allMovieRequest)
	//if handleErr != nil {
	//	fmt.Println(handleErr)
	//}

	//redisClient := newRedisClient()
	//redisErr := addToRedis(movieItems, redisClient)
	//if redisErr != nil {
	//	panic(redisErr)
	//}
	//fmt.Println(getRandomKeysFromRedis(5, redisClient.rdb, redisClient.ctx))
	//
	//mux := http.NewServeMux()
	//
	//mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	//	handleIncomingGetRequest(redisClient, w, r)
	//})
	//mux.HandleFunc("/image", func(w http.ResponseWriter, r *http.Request) {
	//	handleIncomingRequestForImage(mediaBrowser, w, r)
	//})
	//
	//fmt.Println("HTTP connection started on port 8080")
	//httpServeError := http.ListenAndServe(":8080", CorsMiddleware(mux))
	//if httpServeError != nil {
	//	panic(httpServeError)
	//}
}

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
	}
	fmt.Println(allMovies)
	addMoviesErr := rClient.AddItems(allMovies)

	if addMoviesErr != nil {
		panic(addMoviesErr)
	}

	httpErr := createHttpMux(jClient, rClient, hClient)
	if httpErr != nil {
		panic(httpErr)
	}

}

func createJellyfinHttpClient(jClient jellyfin.Client) (jellyfinHttpClient.Client, error) {
	authResponse, authenticationErr := jClient.AuthenticateByName()

	if authenticationErr != nil {
		fmt.Println("Failed to authenticate")
		return nil, authenticationErr
	}

	jHttpClient, err := jellyfinHttpClient.NewClient(jClient, authResponse)
	if err != nil {
		fmt.Println("Failed to create jellyfinHttpClient")
		return nil, err
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
