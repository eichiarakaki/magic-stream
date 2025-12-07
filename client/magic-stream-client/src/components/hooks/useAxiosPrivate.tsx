import axios from "axios";
import useAuth from "./useAuth.tsx";
import { useEffect } from "react";

const apiURL = import.meta.env.VITE_API_BASE_URL;
const axiosPrivate = axios.create({
  baseURL: apiURL,
  withCredentials: true,
});

const useAxiosPrivate = () => {
  const { auth } = useAuth();

  useEffect(() => {
    const requestInterceptor = axiosPrivate.interceptors.request.use(
      (config) => {
        if (auth?.token) {
          config.headers.Authorization = `Bearer ${auth.token}`;
        }
        return config;
      },
      (error) => Promise.reject(error),
    );
    return () => {
      axiosPrivate.interceptors.response.eject(requestInterceptor);
    };
  }, [auth]);
  return axiosPrivate;
};

export default useAxiosPrivate;
