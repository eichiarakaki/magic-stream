import useAxiosPrivate from "../hooks/useAxiosPrivate.tsx";
import { useEffect, useState } from "react";
import Movies from "../movies/Movies.tsx";
import type Movie from "../movie/Movie.tsx";

const Recommended = () => {
  const [movies, setMovies] = useState<Movie[]>([]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");
  const axiosPrivate = useAxiosPrivate();

  useEffect(() => {
    const fetchRecommendedMovies = async () => {
      setLoading(true);
      setMessage("Loading...");

      try {
        const response = await axiosPrivate.get("/recommended-movies");
        setMovies(response.data);
      } catch (error) {
        console.error("Error fetching recommended movies:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchRecommendedMovies();
  }, []);

  return (
    <>
      {loading ? (
        <h2>Loading...</h2>
      ) : (
        <Movies movies={movies} message={message} />
      )}
    </>
  );
};

export default Recommended;
