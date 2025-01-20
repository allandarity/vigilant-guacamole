package service

import (
	"context"
	"go-jellyfin-api/cmd/config"
	"go-jellyfin-api/cmd/model"
	"go-jellyfin-api/cmd/repository"
)

type JellyfinService interface {
	GetRandomMovies(ctx context.Context, noOfMovies int) ([]model.MovieWithImage, error)
}

type jellyfinService struct {
	cfg             config.JellyfinConfiguration
	movieRepository repository.MovieRepository
}

func NewJellyfinService(cfg config.JellyfinConfiguration, m repository.MovieRepository) JellyfinService {
	return &jellyfinService{
		cfg:             cfg,
		movieRepository: m,
	}
}

func (s jellyfinService) GetRandomMovies(ctx context.Context, noOfMovies int) ([]model.MovieWithImage, error) {
	outcome, err := s.movieRepository.GetRandomMovies(ctx, noOfMovies)

	if err != nil {
		return nil, err
	}

	return outcome, nil
}
