CREATE TABLE watchlist
(
    id              serial PRIMARY KEY,
    title           VARCHAR(255) NOT NULL,
    production_year TIMESTAMPTZ  NOT NULL,
    added_date      TIMESTAMPTZ  NOT NULL
)
