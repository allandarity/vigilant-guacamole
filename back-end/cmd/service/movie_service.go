package service

import (
	"context"
	"errors"
	"go-jellyfin-api/cmd/model"
	"go-jellyfin-api/cmd/repository"
)

type MovieService interface {
	GetMovieByName(ctx context.Context, name string) (*model.Movie, error)
	GetRandomMovies(ctx context.Context, numberOfMovies int) ([]model.MovieWithImage, error)
}

type movieService struct {
	repository repository.MovieRepository
}

func NewMovieService(repository repository.MovieRepository) MovieService {
	return &movieService{
		repository: repository,
	}
}

func (m *movieService) GetMovieByName(ctx context.Context, name string) (*model.Movie, error) {
	movie, err := m.repository.GetMovieByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return movie, nil
}

func (m *movieService) GetRandomMovies(ctx context.Context, numberOfMovies int) ([]model.MovieWithImage, error) {
	if numberOfMovies <= 0 {
		return nil, errors.New("Must provide a positive numberOfMovies")
	}

	movies, err := m.repository.GetRandomMovies(ctx, numberOfMovies)
	if err != nil {
		return nil, err
	}

	return movies, nil
}
