package jellyfin

import (
	"errors"
	"fmt"
	"go-jellyfin-api/cmd/model"
	"os"
)

type Client interface {
	BuildMediaBrowserIdentifier() string
	BuildAuthenticationRequest() (model.AuthRequest, error)
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

func (j jellyfin) GetHost() string {
	return j.host
}

func (j jellyfin) BuildAuthenticationRequest() (model.AuthRequest, error) {
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
		Pw:       password,
	}, nil
}

func getAuthCred(credType string) (string, error) {
	cred := os.Getenv(credType)
	if cred == "" {
		return "", errors.New(credType + " is not set")
	}
	return cred, nil
}
