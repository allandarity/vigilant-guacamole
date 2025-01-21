package model

import "time"

type MovieWatchlistPair struct {
	MovieId     int
	WatchlistId int
	AddedDate   time.Time
}
