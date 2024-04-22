import {Show} from "../../model/Show.tsx";
import "./Poster.css"
import {useEffect, useState} from "react";
import {getMoviePoster} from "../../api/MovieInfoApi.ts";

interface PosterProps {
	show: Show
}

function Poster({show}: PosterProps) {

	const [posterImageData, setPosterImageData] = useState<string>();

	useEffect(() => {
		getMoviePoster(show.id).then(output => {
			setPosterImageData(output)
		})
	}, [show.id])


	const handleClick = () => {
		show.onSelect();
	};


	return (
		<>
			{
				posterImageData && (
					<div
						className={`movie-card ${show.isSelected ? 'movie-card--selected' : ''}`}
						onClick={handleClick}
					>
						<div
							className="movie-card__backdrop"
							style={{backgroundImage: `url(${posterImageData})`}}
						>
							<div className="movie-card__text-container">
								<h2 className="movie-card__title">{show.title}</h2>
								<p className="movie-card__details">
									{show.year} | {show.rating}
								</p>
							</div>
						</div>
					</div>
				)
			}
			{
				!posterImageData && (
					<div className="movie-card">
						<div
							className="movie-card__backdrop"
						>
							<div className="movie-card__content">
								<h2 className="movie-card__title">{show.title}</h2>
								<p className="movie-card__details">
									{show.year} | {show.rating}
								</p>
							</div>
						</div>
					</div>
				)
			}
			{show.isSelected && (
				<div className="movie-card__description-box">
					<p>mpv http://192.168.1.157:8096/Videos/{show.id}/stream.mkv</p>
				</div>
			)}
		</>
	)
}

export default Poster;