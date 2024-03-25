package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type mediaBrowser struct {
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

func (m mediaBrowser) getAuthenticationHeader() string {
	return fmt.Sprintf("MediaBrowser client=\"%s\", Device=\"%s\", DeviceId=\"%s\", Version=\"%s\", Token=\"%s\"", m.client, m.device, m.deviceId, m.version, m.token)
}

func makeAuthenticationRequest(mediaBrowser mediaBrowser, authenticationRequest authenticationRequest) (*http.Request, error) {
	jsonBody, err := json.Marshal(authenticationRequest)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", os.Getenv("host")+"/Users/AuthenticateByName", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", mediaBrowser.getAuthenticationHeader())
	req.Header.Set("content-type", "application/json")

	return req, nil
}

func handleAuthenticationRequest(req *http.Request) ([]byte, error) {
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

func main() {
	mediaBrowser := &mediaBrowser{
		client:   "Elliott Jellyfin Launcher",
		device:   "Laptop",
		deviceId: "TW96aWxsYS81LjAgKFgxMTsgTGludXggeDg2XzY0OyBydjo5NC4wKSBHZWNrby8yMDEwMDEwMSBGaXJlZm94Lzk0LjB8MTYzODA1MzA2OTY4Mw11",
		version:  "10.8.8",
		token:    "c9fbf0c11ba84e05ad4d08a78e8cc955",
	}
	authenticationRequest := &authenticationRequest{
		Username: "elliott",
		Pw:       "qqqqqq",
	}
	authRequest, err := makeAuthenticationRequest(*mediaBrowser, *authenticationRequest)
	if err != nil {
		fmt.Println(err)
	}
	authResponseBody, err := handleAuthenticationRequest(authRequest)
	if err != nil {
		fmt.Println(err)
	}
	var authResponse authenticationResponse
	unmarshalErr := json.Unmarshal(authResponseBody, &authResponse)
	if unmarshalErr != nil {
		fmt.Println(err)
	}
	fmt.Println(authResponse)
}
