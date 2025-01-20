package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-jellyfin-api/cmd/config"
	"go-jellyfin-api/cmd/model"
	"io"
	"net/http"
	"strconv"
)

const (
	ImageMaxWidth  = 400
	ImageMaxHeight = 400
)

type Client interface {
	GetMovieFolderParentId() (string, error)
	GetRequest(url string) (*http.Request, error)
	MakeHttpClientRequest(request *http.Request) ([]byte, error)
	GetAllMoviesRequest(parentId string) (model.Items, error)
	AuthenticateByName() error
	PopulateMovieImageData(items model.Items) (*model.Items, error)
}

type jellyfinHttpClient struct {
	authResponse          model.AuthResponse
	jellyfinConfiguration config.JellyfinConfiguration
}

func NewClient(cfg config.JellyfinConfiguration) (Client, error) {
	return &jellyfinHttpClient{
		authResponse:          model.AuthResponse{},
		jellyfinConfiguration: cfg,
	}, nil
}

func (h *jellyfinHttpClient) AuthenticateByName() error {
	requestBody, err := json.Marshal(h.jellyfinConfiguration.BuildAuthenticationRequest())
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.jellyfinConfiguration.GetHost()+"/Users/AuthenticateByName", bytes.NewBuffer(requestBody))
	req.Header.Set("Authorization", h.jellyfinConfiguration.BuildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	if err != nil {
		return err
	}

	resp, err := h.MakeHttpClientRequest(req)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(resp, &h.authResponse); err != nil {
		fmt.Println("can't unmarshal authResponse")
		return err
	}
	return nil
}

func (h *jellyfinHttpClient) GetMovieFolderParentId() (string, error) {
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
		fmt.Println("failed to unmarshal", err)
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

func (h *jellyfinHttpClient) GetRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", h.jellyfinConfiguration.BuildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	return req, nil
}

func (h *jellyfinHttpClient) MakeHttpClientRequest(request *http.Request) ([]byte, error) {
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

func (h *jellyfinHttpClient) getMovieParentIdRequestUrl() string {
	return fmt.Sprintf("%s/Users/%s/Items", h.jellyfinConfiguration.GetHost(), h.authResponse.User.Id)
}

func (h *jellyfinHttpClient) GetAllMoviesRequest(parentId string) (model.Items, error) {
	url := fmt.Sprintf("%s/Users/%s/Items?ParentId=%s", h.jellyfinConfiguration.GetHost(), h.authResponse.User.Id, parentId)
	req, err := h.GetRequest(url)
	if err != nil {
		return model.Items{}, err
	}
	resp, err := h.MakeHttpClientRequest(req)
	if err != nil {
		return model.Items{}, err
	}

	var items model.Items
	if err := json.Unmarshal(resp, &items); err != nil {
		fmt.Println("failed to unmarshal")
		return model.Items{}, err
	}
	return items, nil
}

// TODO: skip this if db is full
func (h *jellyfinHttpClient) PopulateMovieImageData(items model.Items) (*model.Items, error) {
	for i := range items.ItemElements {
		item := &items.ItemElements[i]
		image, err := h.getMovieImageData(item)
		if err != nil {
			return nil, err
		}

		a, _ := strconv.Atoi(item.Id)
		var movieImage model.MovieImage
		movieImage.MovieId = a
		movieImage.ImageData = image
		item.Image = movieImage
	}
	return &items, nil
}

func (h *jellyfinHttpClient) getMovieImageData(item *model.ItemsElement) ([]byte, error) {
	getImageUrl := fmt.Sprintf("%s/Items/%s/Images/Primary?MaxWidth=%d&MaxHeight=%d",
		h.jellyfinConfiguration.GetHost(), item.Id, ImageMaxWidth, ImageMaxHeight)
	req, err := h.GetRequest(getImageUrl)
	if err != nil {
		return nil, fmt.Errorf("error creating request url=%s: %w", getImageUrl, err)
	}
	resp, err := h.MakeHttpClientRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request url=%s: %w", getImageUrl, err)
	}
	return resp, nil
}
