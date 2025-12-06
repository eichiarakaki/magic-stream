import Movie from '../movie/Movie.tsx'


const Movies = ({movies, message}: {movies: [Movie], message: string}) => {
    return (
        <div className={"container mt-4"}>
            <div className={"row"}>
                {movies && movies.length > 0
                ? movies.map((movie) => (
                    <Movie key={movie._id} movie={movie}></Movie>
                    ))
                    : <h2>{message}</h2>
                }
            </div>
        </div>
    )
}
export default Movies;