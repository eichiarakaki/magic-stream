import "./App.css";
import Home from "./components/home/Home.tsx";
import Header from "./components/header/Header.tsx";
import { Route, Routes, useNavigate } from "react-router-dom";
import Register from "./components/register/Register.tsx";
import Login from "./components/login/Login.tsx";
import Layout from "./components/Layout.tsx";
import RequiredAuth from "./components/RequiredAuth.tsx";
import Recommended from "./components/recommended/Recommended.tsx";
import axiosClient from "./api/axiosConfig.ts";
import Review from "./components/review/Review.tsx";
import useAuth from "./components/hooks/useAuth.tsx";

function App() {
  const navigate = useNavigate();
  const { auth, setAuth } = useAuth();

  const updateMovieReview = (imdb_id: string) => {
    navigate(`/review/${imdb_id}`);
  };

  const handleLogout = async () => {
    try {
      const response = await axiosClient.post("/logout", {
        user_id: auth.user_id,
      });
      console.log(response);
      setAuth(null);
      console.log("User logged out");
    } catch (error) {
      console.error("Error logging out: ", error);
    }
  };
  return (
    <>
      <Header handleLogout={handleLogout} />
      <Routes path={"/"} element={<Layout />}>
        <Route
          path={"/"}
          element={<Home updateMovieReview={updateMovieReview} />}
        ></Route>
        <Route path={"/register"} element={<Register />}></Route>
        <Route path={"/login"} element={<Login />}></Route>
        <Route element={<RequiredAuth />}></Route>
        <Route
          path={"/recommended-movies"}
          element={<Recommended updateMovieReview={updateMovieReview} />}
        ></Route>
        <Route path={"/review/:imdb_id"} element={<Review />}></Route>
      </Routes>
    </>
  );
}

export default App;
