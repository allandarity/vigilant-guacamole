import './App.css'
import {useEffect, useState} from "react";
import {getRandomMovies} from "./api/MovieInfoApi.ts";
import {Movie} from "./model/Movies.ts";
import Poster from "./components/poster/Poster.tsx";


function App() {

	const [movieData, setMovieData] = useState<Movie[]>();
	const [loadData, setLoadData] = useState<boolean>(false);
	const [selectedMovieIndex, setSelectedMovieIndex] = useState<number | null>(null);

	useEffect(() => {
		if (!loadData) {
			getRandomMovies().then(data => {
				setMovieData(data)
				setLoadData(true)
			})
		}
	}, [loadData])

	const handleMovieSelect = (index: number) => {
		setSelectedMovieIndex(selectedMovieIndex === index ? null : index);
	};

	return (
		<>
			{movieData ? (
				<div className="movie-cards-container">
					{movieData.map((movie, index) => (
						<div key={movie.id} className="movie-item">
							<Poster show={{
								title: movie.name,
								id: movie.id,
								year: movie.productionYear,
								rating: movie.communityRating,
								isSelected: index === selectedMovieIndex,
								onSelect: () => handleMovieSelect(index),
							}}/>
						</div>
					))}
				</div>
			) : (
				<div>Loading</div>
			)}
		</>
	)
}

export default App
