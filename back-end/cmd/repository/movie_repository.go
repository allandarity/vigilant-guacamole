package repository

import (
	"context"
	"fmt"
	"go-jellyfin-api/cmd/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MovieRepository interface {
	PopulateMovieDatabase(ctx context.Context, items *model.Items) error
	GetMovieByName(ctx context.Context, name string) (*model.Movie, error)
	GetRandomMovies(ctx context.Context, numberOfMovies int) ([]model.MovieWithImage, error)
	GetAllMovies(ctx context.Context) ([]model.Movie, error)
	GetMovieById(ctx context.Context, id int) (*model.Movie, error)
}

type movieRepository struct {
	pool *pgxpool.Pool
}

func NewMovieRepository(pool *pgxpool.Pool) MovieRepository {
	return &movieRepository{
		pool: pool,
	}
}

func (m *movieRepository) GetMovieById(ctx context.Context, id int) (*model.Movie, error) {
	var movie model.Movie
	err := m.pool.QueryRow(
		ctx,
		"SELECT title, production_year, community_rating from movie where id = $1",
		id,
	).Scan(&movie.Name, movie.ProductionYear, movie.CommunityRating)
	if err != nil {
		return nil, err
	}
	return &movie, nil
}

func (m *movieRepository) GetMovieByName(ctx context.Context, name string) (*model.Movie, error) {
	var movie model.Movie
	err := m.pool.QueryRow(
		ctx,
		"SELECT title, production_year, community_rating from movie where name = $1",
		name,
	).Scan(&movie.Name, movie.ProductionYear, movie.CommunityRating)
	if err != nil {
		return nil, nil
	}
	return &movie, nil
}

func (m *movieRepository) PopulateMovieDatabase(ctx context.Context, items *model.Items) error {
	batch := &pgx.Batch{}

	for _, item := range items.ItemElements {
		batch.Queue(
			`INSERT INTO movie (jellyfin_id, title, production_year, community_rating) 
             VALUES ($1, $2, $3, $4) 
             ON CONFLICT (jellyfin_id) DO NOTHING`,
			item.Id,
			item.Name,
			item.ProductionYear,
			item.CommunityRating,
		)

		if item.Image.ImageData != nil {
			batch.Queue(
				`INSERT INTO movie_image (movie_id, image_data) 
                 SELECT id, $2 FROM movie WHERE jellyfin_id = $1
                 ON CONFLICT (movie_id) DO NOTHING`,
				item.Id,
				item.Image.ImageData,
			)
		}
	}

	if batch.Len() > 0 {
		br := m.pool.SendBatch(ctx, batch)
		defer br.Close()
		if err := br.Close(); err != nil {
			return fmt.Errorf("failed to execute batch: %w", err)
		}
	}

	return nil
}

func (m *movieRepository) GetRandomMovies(ctx context.Context, numberOfMovies int) ([]model.MovieWithImage, error) {
	query := `
    SELECT m.id, m.jellyfin_id, m.title, m.production_year, m.community_rating, mi.image_data
    FROM movie m
    LEFT JOIN movie_image mi ON m.id = mi.movie_id
    ORDER BY RANDOM()
    LIMIT $1
  `
	rows, err := m.pool.Query(ctx, query, numberOfMovies)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.MovieWithImage
	for rows.Next() {
		var movie model.MovieWithImage
		err := rows.Scan(
			&movie.Movie.Id,
			&movie.Movie.JellyfinId,
			&movie.Movie.Name,
			&movie.Movie.ProductionYear,
			&movie.Movie.CommunityRating,
			&movie.MovieImage.ImageData,
		)
		if err != nil {
			return nil, err
		}
		movie.MovieImage.MovieId = movie.Movie.Id
		movies = append(movies, movie)
	}
	return movies, nil
}

func (m *movieRepository) GetAllMovies(ctx context.Context) ([]model.Movie, error) {
	query := `
		SELECT id, jellyfin_id, title, production_year, community_rating from movie
	`
	rows, err := m.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []model.Movie
	for rows.Next() {
		var movie model.Movie
		if err := rows.Scan(&movie.Id, &movie.JellyfinId, &movie.Name,
			&movie.ProductionYear, &movie.CommunityRating); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}
