import Movie from "../movie/Movie.tsx";

const Movies = ({
  movies,
  message,
  updateMovieReview,
}: {
  movies: Movie[];
  message: string;
  updateMovieReview: (imdb_id: string) => void;
}) => {
  return (
    <div className={"container mt-4"}>
      <div className={"row"}>
        {movies && movies.length > 0 ? (
          movies.map((movie) => (
            <Movie
              key={movie._id}
              movie={movie}
              updateMovieReview={updateMovieReview}
            ></Movie>
          ))
        ) : (
          <h2>{message}</h2>
        )}
      </div>
    </div>
  );
};
export default Movies;
