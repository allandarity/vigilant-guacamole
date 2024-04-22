package jellyfin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go-jellyfin-api/pkg/model"
	"net/http"
	"os"
)

type Client interface {
	BuildMediaBrowserIdentifier() string
	AuthenticateByName() (model.AuthResponse, error)
	GetHost() string
}

type jellyfin struct {
	host     string
	client   string
	device   string
	deviceId string
	version  string
	token    string
}

func NewClient() (Client, error) {
	token := os.Getenv("DEVICE_TOKEN")
	if token == "" {
		return jellyfin{}, errors.New("DEVICE_TOKEN is not set")
	}
	deviceId := os.Getenv("DEVICE_ID")
	if deviceId == "" {
		return jellyfin{}, errors.New("DEVICE_ID is not set")
	}
	host := os.Getenv("JELLYFIN_HOST")
	if host == "" {
		return jellyfin{}, errors.New("JELLYFIN_HOST is not set")
	}
	return jellyfin{
		client:   "Elliott Jellyfin Launcher",
		device:   "Laptop",
		deviceId: deviceId,
		version:  "10.8.8",
		token:    token,
		host:     host,
	}, nil
}

func (j jellyfin) BuildMediaBrowserIdentifier() string {
	return fmt.Sprintf("MediaBrowser client=\"%s\", Device=\"%s\", DeviceId=\"%s\", Version=\"%s\", Token=\"%s\"", j.client, j.device, j.deviceId, j.version, j.token)
}

func (j jellyfin) AuthenticateByName() (model.AuthResponse, error) {
	authenticationRequest, err := buildAuthenticationRequest()
	if err != nil {
		return model.AuthResponse{}, err
	}
	requestBody, err := json.Marshal(authenticationRequest)
	if err != nil {
		return model.AuthResponse{}, err
	}
	req, err := http.NewRequest("POST", j.host+"/Users/AuthenticateByName", bytes.NewBuffer(requestBody))
	if err != nil {
		return model.AuthResponse{}, err
	}

	req.Header.Set("Authorization", j.BuildMediaBrowserIdentifier())
	req.Header.Set("content-type", "application/json")

	//TODO we dont actually even authenticate at any point

	return model.AuthResponse{
		User: model.AuthUser{
			Name: "",
			Id:   "",
		},
		Token: "",
	}, nil
}

func (j jellyfin) GetHost() string {
	return j.host
}

func buildAuthenticationRequest() (model.AuthRequest, error) {
	username, err := getAuthCred("USERNAME")
	if err != nil {
		return model.AuthRequest{}, err
	}
	password, err := getAuthCred("PASSWORD")
	if err != nil {
		return model.AuthRequest{}, err
	}
	return model.AuthRequest{
		Username: username,
		Password: password,
	}, nil
}

func getAuthCred(credType string) (string, error) {
	cred := os.Getenv(credType)
	if cred == "" {
		return "", errors.New(credType + " is not set")
	}
	return cred, nil
}
