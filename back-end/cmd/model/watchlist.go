package model

type Watchlist struct {
	Items []WatchlistItem
}

type WatchlistItem struct {
	Title        string
	DateAdded    string
	DateReleased string
}
