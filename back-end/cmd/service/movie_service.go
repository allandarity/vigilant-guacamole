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
	GetAllMovies(ctx context.Context) ([]model.Movie, error)
	GetMovieById(ctx context.Context, id int) (model.Movie, error)
	GetMovieByIdWithImage(ctx context.Context, id int) (model.MovieWithImage, error)
}

type movieService struct {
	repository repository.MovieRepository
}

func NewMovieService(repository repository.MovieRepository) MovieService {
	return &movieService{
		repository: repository,
	}
}

func (m *movieService) GetMovieById(ctx context.Context, id int) (model.Movie, error) {
	movie, err := m.repository.GetMovieById(ctx, id)
	if err != nil {
		return model.Movie{}, err
	}
	return movie, err
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

func (m *movieService) GetAllMovies(ctx context.Context) ([]model.Movie, error) {
	movies, err := m.repository.GetAllMovies(ctx)
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (m *movieService) GetMovieByIdWithImage(ctx context.Context, id int) (model.MovieWithImage, error) {
	movie, err := m.repository.GetMovieByIdWithImage(ctx, id)
	if err != nil {
		return model.MovieWithImage{}, err
	}
	return movie, nil
}
