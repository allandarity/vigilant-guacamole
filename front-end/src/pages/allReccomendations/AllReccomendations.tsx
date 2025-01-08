import './AllRecommendations.css'
import { useEffect, useState } from "react";
import { getRandomMovies } from "../../api/MovieInfoApi.ts";
import { Movie } from "../../model/Movies.ts";
import Poster from "../../components/poster/Poster.tsx";


function AllRecommendations() {

  const [movieData, setMovieData] = useState<Movie[]>();
  const [loadData, setLoadData] = useState<boolean>(false);
  const [selectedMovieIndex, setSelectedMovieIndex] = useState<number | null>(null);

  useEffect(() => {
    if (!loadData) {
      setLoadData(true)
      getRandomMovies().then(data => {
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
              }} />
            </div>
          ))}
        </div>
      ) : (
        <div>Loading</div>
      )}
    </>
  )
}

export default AllRecommendations
