package http

import (
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
	GetAllMoviesRequest(parentId string) (model.Items, error)
}

type jellyfinHttpClient struct {
	jellyfin     jellyfin.Client
	authResponse model.AuthResponse
}

func (h jellyfinHttpClient) GetMovieFolderParentId() (string, error) {
	httpRequest, err := h.GetRequest(h.getMovieParentIdRequestUrl())
	if err != nil {
		return "", err
	}
	httpResponse, err := h.MakeHttpClientRequest(httpRequest)
	if err != nil {
		return "", err
	}

	var items model.Items
	unmarshalErr := unmarshalForType(httpResponse, &items)
	if unmarshalErr != nil {
		return "", unmarshalErr
	}
	movies := "Movies"
	collection := items.GetItemByName(movies)
	if collection.IsEmpty() {
		return "", errors.New("unable to find the Movies collection")
	}
	if collection.IsOfCorrectType(movies) {
		return "", fmt.Errorf("the collection of the wrong type - wasnt %s", movies)
	}
	fmt.Println(collection)
	return collection.Id, nil
}

func NewClient(jellyfin jellyfin.Client, authResponse model.AuthResponse) (Client, error) {
	return jellyfinHttpClient{
		jellyfin:     jellyfin,
		authResponse: authResponse,
	}, nil
}

func (h jellyfinHttpClient) GetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", h.jellyfin.GetHost())
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

func (h jellyfinHttpClient) GetAllMoviesRequest(parentId string) (model.Items, error) {
	url := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s", h.jellyfin.GetHost(), h.authResponse.User.Id, parentId)
	fmt.Println(url)
	req, err := h.GetRequest(url)
	if err != nil {
		return model.Items{}, err
	}
	resp, err := h.MakeHttpClientRequest(req)
	if err != nil {
		return model.Items{}, err
	}

	var items model.Items
	unmarshalErr := unmarshalForType(resp, &items)
	if unmarshalErr != nil {
		return model.Items{}, unmarshalErr
	}
	return items, nil
}

func unmarshalForType[T any](response []byte, target *T) error {
	err := json.Unmarshal(response, target)
	if err != nil {
		return err
	}
	return nil
}
