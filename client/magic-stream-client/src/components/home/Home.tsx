import { useState, useEffect } from "react";
import axiosClient from "../../api/axiosConfig.ts";
import Movies from "../movies/Movies.tsx";
import Movie from "../movie/Movie.tsx";

const Home = ({ updateMovieReview }) => {
  const [movies, setMovies] = useState([Movie]);
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    const fetchMovies = async () => {
      setLoading(true);
      setMessage("");

      try {
        const response = await axiosClient.get("/movies");
        setMovies(response.data);
        if (response.data.length === 0) {
          setMessage("There are currently no movies available!");
        }
      } catch (e) {
        console.error(e);
        setMessage("Error fetching movies");
      } finally {
        setLoading(false);
      }
    };
    fetchMovies();
  }, []);

  return (
    <>
      {loading ? (
        <h2>Loading...</h2>
      ) : (
        <Movies
          movies={movies}
          updateMovieReview={updateMovieReview}
          message={message}
        ></Movies>
      )}
    </>
  );
};

export default Home;
