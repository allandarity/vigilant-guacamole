CREATE TABLE movie(
  id serial PRIMARY KEY,
  jellyfin_id VARCHAR(255) NOT NULL UNIQUE,
  title VARCHAR(255) NOT NULL,
  production_year SMALLINT NOT NULL,
  community_rating DECIMAL not null
)
