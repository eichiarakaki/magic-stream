import { useLocation, Navigate, Outlet } from "react-router-dom";
import useAuth from "./hooks/useAuth.tsx";

const RequiredAuth = () => {
  const { auth } = useAuth();
  const location = useLocation();

  return auth ? (
    <Outlet />
  ) : (
    <Navigate to="login" state={{ from: location }} replace></Navigate>
  );
};

export default RequiredAuth;
