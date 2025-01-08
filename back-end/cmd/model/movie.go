package model

type MovieWithImage struct {
	Movie      Movie
	MovieImage MovieImage
}

type Movie struct {
	Id              int
	JellyfinId      string
	Name            string
	ProductionYear  int16
	CommunityRating float32
}

type MovieImage struct {
	MovieId   int
	ImageData []byte
}
