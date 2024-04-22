import {Movie} from "../model/Movies.ts";

const BASE_URL = "http://localhost:8080"

export async function getRandomMovies(): Promise<Movie[]> {
	const response = await fetch(`${BASE_URL}/`)
	const data = await response.json()
	return data.map(convertToMovie)
}

export async function getMoviePoster(id: string) {
	const response = await fetch(`${BASE_URL}/image?id=${id}`)
	const data =  await response.blob()
	return URL.createObjectURL(data);
}

function convertToMovie(data: any): Movie {
	return {
		productionYear: data.ProductionYear,
		communityRating: data.CommunityRating,
		name: data.Name,
		id: data.Id
	}
}