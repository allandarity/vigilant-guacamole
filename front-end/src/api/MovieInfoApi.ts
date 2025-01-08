import {Movie} from "../model/Movies.ts";

const BASE_URL = "http://localhost:8080"

export async function getRandomMovies(): Promise<Movie[]> {
	const response = await fetch(`${BASE_URL}/movies/random`)
	const data = await response.json()
	return data.map(convertToMovie)
}

function convertToMovie(data: any): Movie {
  const binaryImageData = Uint8Array.from(atob(data.MovieImage.ImageData), char => char.charCodeAt(0));
	return {
    movieId: data.Movie.Id,
		productionYear: data.Movie.ProductionYear,
		communityRating: data.Movie.CommunityRating,
		name: data.Movie.Name,
		jellyfinId: data.Movie.JellyfinId,
    imageData: binaryImageData
	}
}
