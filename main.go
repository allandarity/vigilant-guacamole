package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	Items []ItemsElement `json:"Items"`
}

type ItemsElement struct {
	Name string `json:"Name"`
	Id   string `json:"Id"`
	Type string `json:"Type"`
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

func makeGetMovieParentIdRequest(mediaBrowser mediaBrowser, authResponse authenticationResponse) (*http.Request, error) {
	url := fmt.Sprintf("%s/Users/%s/Items", mediaBrowser.host, authResponse.User.Id)
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

func handleGetMovieParentIdRequest(req *http.Request) (Items, error) {
	respBody, err := makeHttpClientRequest(req)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return Items{}, err
	}
	var items Items
	unmarshalErr := json.Unmarshal(respBody, &items)
	if unmarshalErr != nil {
		return Items{}, unmarshalErr
	}
	return items, nil
}

func handleMakeAuthenticationRequest(req *http.Request) (authenticationResponse, error) {
	authResponseBody, err := makeHttpClientRequest(req)
	if err != nil {
		fmt.Println(err)
		return authenticationResponse{}, err
	}
	var authResponse authenticationResponse
	unmarshalErr := json.Unmarshal(authResponseBody, &authResponse)
	if unmarshalErr != nil {
		fmt.Println(err)
		return authenticationResponse{}, err
	}
	return authResponse, nil
}

func getItemByName(items Items, name string) ItemsElement {
	for _, item := range items.Items {
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

	parentFolderRequest, err := makeGetMovieParentIdRequest(mediaBrowser, authResponse)
	if err != nil {
		fmt.Println(err)
	}
	parentFolderResponse, err := handleGetMovieParentIdRequest(parentFolderRequest)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(parentFolderResponse.Items)
	fmt.Println(getMoviesParentId(parentFolderResponse))
}
