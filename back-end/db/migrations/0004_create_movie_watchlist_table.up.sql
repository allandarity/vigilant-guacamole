CREATE TABLE movie_watchlist(
  movie_id INTEGER REFERENCES movie(id),
  watchlist_id INTEGER REFERENCES watchlist(id),
  added_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY_KEY(movie_id, watchlist_id)
);

CREATE INDEX idx_movie_watchlist_movie ON movie_watchlist(movie_id)
CREATE INDEX idx_movie_watchlist_watchlist ON movie_watchlist(watchlist_id)


