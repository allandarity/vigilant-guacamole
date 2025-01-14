package service

import "go-jellyfin-api/cmd/config"

type JellyfinService interface{}

type jellyfinService struct {
	cfg config.JellyfinConfiguration
}

func NewJellyfinService(cfg config.JellyfinConfiguration) JellyfinService {
	return &jellyfinService{
		cfg: cfg,
	}
}
