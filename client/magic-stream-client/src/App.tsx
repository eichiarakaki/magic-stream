import "./App.css";
import Home from "./components/home/Home.tsx";
import Header from "./components/header/Header.tsx";
import { Route, Routes, useNavigate } from "react-router-dom";
import Register from "./components/register/Register.tsx";
import Login from "./components/login/Login.tsx";
import Layout from "./components/Layout.tsx";
import RequiredAuth from "./components/RequiredAuth.tsx";
import Recommended from "./components/recommended/Recommended.tsx";

function App() {
  return (
    <>
      <Header />
      <Routes path={"/"} element={<Layout />}>
        <Route path={"/"} element={<Home />}></Route>
        <Route path={"/register"} element={<Register />}></Route>
        <Route path={"/login"} element={<Login />}></Route>
        <Route element={<RequiredAuth />}></Route>
        <Route path={"/recommended-movies"} element={<Recommended />}></Route>
      </Routes>
    </>
  );
}

export default App;
