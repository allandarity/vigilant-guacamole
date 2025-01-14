package config

import (
	"fmt"
	"go-jellyfin-api/cmd/model"
	"os"
)

type JellyfinConfiguration interface {
	BuildMediaBrowserIdentifier() string
	GetHost() string
	BuildAuthenticationRequest() model.AuthRequest
}

type jellyfinConfiguration struct {
	host     string
	client   string
	device   string
	deviceId string
	version  string
	token    string
	hostEnvs hostEnvs
}

type hostEnvs struct {
	deviceToken  string
	deviceId     string
	jellyfinHost string
	username     string
	password     string
}

func NewJellyfinConfiguration() (JellyfinConfiguration, error) {
	envs, err := validateEnv()
	if err != nil {
		return nil, err
	}
	return &jellyfinConfiguration{
		host:     envs.jellyfinHost,
		client:   "JFin Launcher",
		device:   "Laptop",
		deviceId: envs.deviceId,
		version:  "10.8.8",
		token:    envs.deviceToken,
		hostEnvs: *envs,
	}, err
}

func (j *jellyfinConfiguration) BuildMediaBrowserIdentifier() string {
	return fmt.Sprintf("MediaBrowser client=\"%s\", Device=\"%s\", DeviceId=\"%s\", Version=\"%s\", Token=\"%s\"", j.client, j.device, j.deviceId, j.version, j.token)
}

func (j *jellyfinConfiguration) GetHost() string {
	return j.host
}

func (j *jellyfinConfiguration) BuildAuthenticationRequest() model.AuthRequest {
	return model.AuthRequest{
		Username: j.hostEnvs.username,
		Pw:       j.hostEnvs.password,
	}
}

var requiredEnvs = []string{
	"DEVICE_ID",
	"DEVICE_TOKEN",
	"JELLYFIN_HOST",
	"USERNAME",
	"PASSWORD",
}

func validateEnv() (*hostEnvs, error) {
	envs := &hostEnvs{}
	for _, key := range requiredEnvs {
		value := os.Getenv(key)
		if value == "" {
			return nil, fmt.Errorf("value of key %s does not exist", key)
		}

		switch key {
		case "DEVICE_ID":
			envs.deviceId = value
		case "DEVICE_TOKEN":
			envs.deviceToken = value
		case "JELLYFIN_HOST":
			envs.jellyfinHost = value
		case "USERNAME":
			envs.username = value
		case "PASSWORD":
			envs.password = value
		}
	}
	return envs, nil
}
