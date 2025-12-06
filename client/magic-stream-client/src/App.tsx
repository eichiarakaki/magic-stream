import "./App.css";
import Home from "./components/home/Home.tsx";
import Header from "./components/header/Header.tsx";
import { Route, Routes, useNavigate } from "react-router-dom";
import Register from "./components/register/Register.tsx";
import Login from "./components/login/Login.tsx";

function App() {
  return (
    <>
      <Header />
      <Routes>
        <Route path={"/"} element={<Home />}></Route>
        <Route path={"/register"} element={<Register />}></Route>
        <Route path={"/Login"} element={<Login />}></Route>
      </Routes>
    </>
  );
}

export default App;
