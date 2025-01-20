package service

import (
	"context"
	"go-jellyfin-api/cmd/model"
	"go-jellyfin-api/cmd/repository"
	"strings"
)

type MovieWatchlistService interface {
	PopulateDatabase(ctx context.Context) ([]model.MovieWatchlistPair, error)
}

type movieWatchlistService struct {
	mService MovieService
	wService WatchlistService
	repo     repository.MovieWatchlistRepository
}

func NewMovieWatchlistService(mService MovieService, wService WatchlistService, repo repository.MovieWatchlistRepository) MovieWatchlistService {
	return &movieWatchlistService{
		mService: mService,
		wService: wService,
		repo:     repo,
	}
}

func (mw *movieWatchlistService) PopulateDatabase(ctx context.Context) ([]model.MovieWatchlistPair, error) {
	movies, err := mw.mService.GetAllMovies(ctx)
	if err != nil {
		return nil, err
	}

	watchlist, err := mw.wService.GetAllWatchlist(ctx)
	if err != nil {
		return nil, err
	}

	var pairs []model.MovieWatchlistPair
	for _, movie := range movies {
		for _, watchlist := range watchlist {
			if mw.matches(movie, watchlist) {
				pairs = append(pairs, model.MovieWatchlistPair{
					MovieId:     movie.Id,
					WatchlistId: watchlist.Id,
					AddedDate:   watchlist.DateAdded,
				})
			}
		}
	}
	err = mw.repo.InsertPairs(ctx, pairs)
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (mw *movieWatchlistService) matches(movie model.Movie, watchlist model.WatchlistItem) bool {
	return strings.EqualFold(movie.Name, watchlist.Title) && int(movie.ProductionYear) == watchlist.DateReleased.Year()
}
