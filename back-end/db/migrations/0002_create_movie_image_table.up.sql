CREATE TABLE movie_image(
  id serial PRIMARY KEY,
  movie_id INT NOT NULL UNIQUE,
  image_data BYTEA NOT NULL,

  FOREIGN KEY (movie_id) REFERENCES movie(id) ON DELETE CASCADE
);

CREATE INDEX idx_movie_image_movie_id ON movie_image(movie_id);
