import { Form, Button } from "react-bootstrap";
import { useRef, useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import useAxiosPrivate from "../hooks/useAxiosPrivate.tsx";
import useAuth from "../hooks/useAuth";
import Movie from "../movie/Movie";
import Spinner from "../spinner/Spinner";
import * as React from "react";

const Review = () => {
  const [movie, setMovie] = useState<Movie | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const revText = useRef<HTMLTextAreaElement>(null);
  const { imdb_id } = useParams();
  const { auth } = useAuth();
  const axiosPrivate = useAxiosPrivate();

  useEffect(() => {
    const fetchMovie = async () => {
      setLoading(true);
      try {
        const response = await axiosPrivate.get(`/movie/${imdb_id}`);
        setMovie(response.data);
      } catch (error) {
        console.error("Error fetching movie:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchMovie();
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!revText.current) return;

    setLoading(true);
    try {
      const response = await axiosPrivate.patch(`/update-review/${imdb_id}`, {
        admin_review: revText.current.value,
      });

      setMovie((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          admin_review: response.data?.admin_review ?? prev.admin_review,
          ranking: {
            ...prev.ranking,
            ranking_name:
              response.data?.ranking_name ?? prev.ranking?.ranking_name,
          },
        };
      });
    } catch (err: any) {
      if (err.response?.status === 401) {
        console.error("Unauthorized - redirecting");
        localStorage.removeItem("user");
      } else {
        console.error("Error updating review:", err);
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      {loading ? (
        <Spinner />
      ) : (
        <div className="container py-5">
          <h2 className="text-center mb-4">Admin Review</h2>
          <div className="row justify-content-center">
            <div className="col-12 col-md-6">
              {movie && (
                <div className="shadow rounded p-3 bg-white">
                  <Movie movie={movie} />
                </div>
              )}
            </div>

            <div className="col-12 col-md-6">
              <div className="shadow rounded p-4 bg-light">
                {auth?.role === "ADMIN" ? (
                  <Form onSubmit={handleSubmit}>
                    <Form.Group className="mb-3">
                      <Form.Label>Admin Review</Form.Label>
                      <Form.Control
                        ref={revText}
                        required
                        as="textarea"
                        rows={8}
                        defaultValue={movie?.admin_review}
                        placeholder="Write your review here..."
                      />
                    </Form.Group>
                    <Button variant="info" type="submit">
                      Submit Review
                    </Button>
                  </Form>
                ) : (
                  <div className="alert alert-info">{movie?.admin_review}</div>
                )}
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  );
};

export default Review;
