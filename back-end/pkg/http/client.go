package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-jellyfin-api/pkg/jellyfin"
	"go-jellyfin-api/pkg/model"
	"io"
	"net/http"
)

type Client interface {
	GetMovieFolderParentId() (string, error)
	GetRequest(url string) (*http.Request, error)
	MakeHttpClientRequest(request *http.Request) ([]byte, error)
	GetAllMoviesRequest(parentId string) (*model.Items, error)
	AuthenticateByName() error
}

type jellyfinHttpClient struct {
	jellyfin     jellyfin.Client
	authResponse model.AuthResponse
}

func NewClient(jellyfin jellyfin.Client) (Client, error) {
	return &jellyfinHttpClient{
		jellyfin:     jellyfin,
		authResponse: model.AuthResponse{},
	}, nil
}

func (h *jellyfinHttpClient) AuthenticateByName() error {
	authRequest, err := h.jellyfin.BuildAuthenticationRequest()
	if err != nil {
		panic(err)
	}

	requestBody, err := json.Marshal(authRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.jellyfin.GetHost()+"/Users/AuthenticateByName", bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", h.jellyfin.BuildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	if err != nil {
		return err
	}

	resp, err := h.MakeHttpClientRequest(req)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, &h.authResponse); err != nil {
		fmt.Println("can't unmarshal")
		return err
	}

	return nil
}

func (h jellyfinHttpClient) GetMovieFolderParentId() (string, error) {
	httpRequest, err := h.GetRequest(h.getMovieParentIdRequestUrl())
	if err != nil {
		fmt.Println("Failed to make request to " + h.getMovieParentIdRequestUrl())
		return "", err
	}
	httpResponse, err := h.MakeHttpClientRequest(httpRequest)
	if err != nil {
		fmt.Println("Failed to make http client request")
		return "", err
	}

	var items model.Items
	if err := json.Unmarshal(httpResponse, &items); err != nil {
		fmt.Println("failed to unmarshal")
		return "", err
	}

	movies := "Movies"
	collection := items.GetItemByName(movies)
	if collection.IsEmpty() {
		return "", errors.New("unable to find the Movies collection")
	}
	if collection.IsOfCorrectType(movies) {
		return "", fmt.Errorf("the collection of the wrong type - wasnt %s", movies)
	}
	return collection.Id, nil
}

func (h jellyfinHttpClient) GetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", h.jellyfin.BuildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	return req, nil
}

func (h jellyfinHttpClient) MakeHttpClientRequest(request *http.Request) ([]byte, error) {
	client := &http.Client{}
	resp, err := client.Do(request)
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

func (h jellyfinHttpClient) getMovieParentIdRequestUrl() string {
	return fmt.Sprintf("%s/Users/%s/Items", h.jellyfin.GetHost(), h.authResponse.User.Id)
}

func (h jellyfinHttpClient) GetAllMoviesRequest(parentId string) (*model.Items, error) {
	url := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s", h.jellyfin.GetHost(), h.authResponse.User.Id, parentId)
	req, err := h.GetRequest(url)
	if err != nil {
		return nil, err
	}
	resp, err := h.MakeHttpClientRequest(req)
	if err != nil {
		return nil, err
	}

	var items *model.Items
	if err := json.Unmarshal(resp, &items); err != nil {
		fmt.Println("failed to unmarshal")
		return nil, err
	}
	return items, nil
}
