import { useLocation, Navigate, Outlet } from "react-router-dom";
import { useState } from "react";

const RequiredAuth = () => {
  const [auth] = useState<any>(null);
  const location = useLocation();

  return auth ? (
    <Outlet />
  ) : (
    <Navigate to="login" state={{ from: location }} replace></Navigate>
  );
};

export default RequiredAuth;
