package service

import (
	"context"
	"errors"
	"fmt"
	"go-jellyfin-api/cmd/model"
	"go-jellyfin-api/cmd/repository"
)

type MovieService interface {
	GetMovieByName(ctx context.Context, name string) (*model.Movie, error)
	GetRandomMovies(ctx context.Context, numberOfMovies int) ([]model.MovieWithImage, error)
	GetAllMovies(ctx context.Context) ([]model.Movie, error)
	GetMovieById(ctx context.Context, id int) *model.Movie
}

type movieService struct {
	repository repository.MovieRepository
}

func NewMovieService(repository repository.MovieRepository) MovieService {
	return &movieService{
		repository: repository,
	}
}

func (m *movieService) GetMovieById(ctx context.Context, id int) *model.Movie {
	movie, err := m.repository.GetMovieById(ctx, id)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return movie
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
