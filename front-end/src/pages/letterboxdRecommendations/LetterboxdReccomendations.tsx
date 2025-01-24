import './LetterboxdReccomendations.css'
import {useEffect, useState} from "react";
import {getRandomWatchlistMovies} from "../../api/MovieInfoApi.ts";
import Poster from "../../components/poster/Poster.tsx";
import {Movie} from "../../model/Movies.tsx";

function LetterboxdRecommendations() {
    const [movieData, setMovieData] = useState<Movie[]>();
    const [loadData, setLoadData] = useState<boolean>(false);
    const [selectedMovieIndex, setSelectedMovieIndex] = useState<number | null>(null);

    useEffect(() => {
        if (!loadData) {
            setLoadData(true)
            getRandomWatchlistMovies().then(data => {
                setMovieData(data)
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
                        <div key={movie.movieId} className="movie-item">
                            <Poster show={{
                                title: movie.name,
                                movieId: movie.movieId,
                                jellyfinId: movie.jellyfinId,
                                year: movie.productionYear,
                                rating: movie.communityRating,
                                isSelected: index === selectedMovieIndex,
                                imageData: movie.imageData,
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

export default LetterboxdRecommendations;
