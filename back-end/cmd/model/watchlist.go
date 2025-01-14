package model

import "time"

type Watchlist struct {
	WatchlistItems []WatchlistItem
}

type WatchlistItem struct {
	Title        string
	DateAdded    time.Time
	DateReleased time.Time
}
