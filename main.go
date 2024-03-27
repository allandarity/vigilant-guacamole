package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type mediaBrowser struct {
	host     string
	client   string
	device   string
	deviceId string
	version  string
	token    string
}

type authenticationRequest struct {
	Username string `json:"Username"`
	Pw       string `json:"Pw"`
}

type authenticationResponse struct {
	User  authenticationResponseUser `json:"User"`
	Token string                     `json:"AccessToken"`
}

type authenticationResponseUser struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
}

type getItems struct {
	userId                 string
	parentId               string
	minCommunityRating     int
	enableTotalRecordCount bool
	limit                  int
}

type Items struct {
	ItemElements []ItemsElement `json:"Items"`
}

type ItemsElement struct {
	Name            string  `json:"Name"`
	Id              string  `json:"Id"`
	Type            string  `json:"Type"`
	ProductionYear  int16   `json:"ProductionYear"`
	CommunityRating float32 `json:"CommunityRating"`
}

func (ie ItemsElement) isEmpty() bool {
	return ie.Name == "" || ie.Id == "" || ie.Type == ""
}

func (ie ItemsElement) isOfCorrectType(expectedType string) bool {
	return ie.Type == expectedType
}

func (m mediaBrowser) buildMediaBrowserIdentifier() string {
	return fmt.Sprintf("MediaBrowser client=\"%s\", Device=\"%s\", DeviceId=\"%s\", Version=\"%s\", Token=\"%s\"", m.client, m.device, m.deviceId, m.version, m.token)
}

func makeAuthenticationRequest(mediaBrowser mediaBrowser, authenticationRequest authenticationRequest) (*http.Request, error) {
	jsonBody, err := json.Marshal(authenticationRequest)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", mediaBrowser.host+"/Users/AuthenticateByName", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", mediaBrowser.buildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	return req, nil
}

func buildMediaBrowser() (mediaBrowser, error) {
	token := os.Getenv("DEVICE_TOKEN")
	if token == "" {
		return mediaBrowser{}, errors.New("DEVICE_TOKEN is not set")
	}
	deviceId := os.Getenv("DEVICE_ID")
	if deviceId == "" {
		return mediaBrowser{}, errors.New("DEVICE_ID is not set")
	}
	host := os.Getenv("JELLYFIN_HOST")
	if host == "" {
		return mediaBrowser{}, errors.New("JELLYFIN_HOST is not set")
	}
	return mediaBrowser{
		client:   "Elliott Jellyfin Launcher",
		device:   "Laptop",
		deviceId: deviceId,
		version:  "10.8.8",
		token:    token,
		host:     host,
	}, nil
}

func buildAuthenticationRequest() (authenticationRequest, error) {
	username := os.Getenv("USERNAME")
	if username == "" {
		return authenticationRequest{}, errors.New("USERNAME is not set")
	}
	password := os.Getenv("PASSWORD")
	if password == "" {
		return authenticationRequest{}, errors.New("PASSWORD is not set")
	}
	return authenticationRequest{
		Username: username,
		Pw:       password,
	}, nil
}

func makeGetRequest(url string, mediaBrowser mediaBrowser) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", mediaBrowser.buildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	return req, nil
}

func makeHttpClientRequest(req *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}
	return respBody, nil
}

func makeRequest[T any](req *http.Request, target *T) error {
	resp, err := makeHttpClientRequest(req)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	err = json.Unmarshal(resp, target)
	if err != nil {
		return err
	}
	return nil
}

func handleGetMovieParentIdRequest(req *http.Request) (Items, error) {
	var items Items
	err := makeRequest(req, &items)
	if err != nil {
		return Items{}, err
	}
	return items, nil
}

func handleMakeAuthenticationRequest(req *http.Request) (authenticationResponse, error) {
	var authResponse authenticationResponse
	err := makeRequest(req, &authResponse)
	if err != nil {
		return authenticationResponse{}, err
	}
	return authResponse, nil
}

func getItemByName(items Items, name string) ItemsElement {
	for _, item := range items.ItemElements {
		if item.Name == name {
			return item
		}
	}
	return ItemsElement{}
}

func getMoviesParentId(items Items) (string, error) {
	movies := "Movies"
	collection := getItemByName(items, movies)
	if collection.isEmpty() {
		return "", errors.New("unable to find the Movies collection")
	}
	if collection.isOfCorrectType(movies) {
		return "", fmt.Errorf("the collection of the wrong type - wasnt %s", movies)
	}
	return collection.Id, nil
}

func addToRedis(items Items, rdb *redis.Client, ctx context.Context) error {
	pipe := rdb.Pipeline()
	for _, i := range items.ItemElements {
		key := fmt.Sprintf("movie:%s", i.Id)
		structBytes, err := json.Marshal(i)
		if err != nil {
			fmt.Println(err)
			continue
		}
		pipe.Set(ctx, key, structBytes, 0)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func getRandomKeysFromRedis(n int, rdb *redis.Client, ctx context.Context) ([]ItemsElement, error) {
	var items []ItemsElement
	for i := 0; i < n; i++ {
		item, err := rdb.RandomKey(ctx).Result()
		if err != nil {
			fmt.Println("failed to get randomkey")
			return nil, err
		}
		unmarshalledItem, err := getItemFromRedis(item, rdb, ctx)
		if err != nil {
			fmt.Println("Failed to unmarshal")
			fmt.Println(item)
			return nil, err
		}
		items = append(items, unmarshalledItem)
	}
	return items, nil
}

func getItemFromRedis(key string, rdb *redis.Client, ctx context.Context) (ItemsElement, error) {
	item, err := rdb.Get(ctx, key).Result()
	if err != nil {
		fmt.Println("couldnt get key")
		return ItemsElement{}, err
	}
	var itemElement ItemsElement
	jsonErr := json.Unmarshal([]byte(item), &itemElement)
	if jsonErr != nil {
		fmt.Println("failed to get key from redis")
		fmt.Println(item)
		return ItemsElement{}, jsonErr
	}
	return itemElement, nil
}

func main() {
	mediaBrowser, err := buildMediaBrowser()
	if err != nil {
		panic(err)
	}
	authenticationRequest, err := buildAuthenticationRequest()
	if err != nil {
		panic(err)
	}
	authRequest, err := makeAuthenticationRequest(mediaBrowser, authenticationRequest)
	if err != nil {
		fmt.Println(err)
	}
	authResponse, err := handleMakeAuthenticationRequest(authRequest)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(authResponse.User.Id)
	fmt.Println(authResponse.User.Name)
	fmt.Println(authResponse.Token)

	getMovieParentIdRequestUrl := fmt.Sprintf("%s/Users/%s/Items", mediaBrowser.host, authResponse.User.Id)
	parentFolderRequest, err := makeGetRequest(getMovieParentIdRequestUrl, mediaBrowser)
	if err != nil {
		fmt.Println(err)
	}
	parentFolderResponse, err := handleGetMovieParentIdRequest(parentFolderRequest)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(parentFolderResponse.ItemElements)
	parentId, err := getMoviesParentId(parentFolderResponse)
	if err != nil {
		fmt.Println(err)
	}

	getAllMoviesRequestUrl := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s", mediaBrowser.host, authResponse.User.Id, parentId)
	allMovieRequest, err := makeGetRequest(getAllMoviesRequestUrl, mediaBrowser)
	if err != nil {
		fmt.Println(err)
	}
	movieItems, handleErr := handleGetMovieParentIdRequest(allMovieRequest)
	if handleErr != nil {
		fmt.Println(handleErr)
	}

	fmt.Println("0")

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	fmt.Println("1")
	redisErr := addToRedis(movieItems, rdb, ctx)
	if redisErr != nil {
		panic(redisErr)
	}
	fmt.Println(getRandomKeysFromRedis(5, rdb, ctx))
}
