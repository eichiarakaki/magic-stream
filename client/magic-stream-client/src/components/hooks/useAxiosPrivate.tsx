import axios from "axios";
import { useEffect } from "react";

const apiURL =
  import.meta.env.VITE_API_BASE_URL ||
  import.meta.env.VITE_API_URL ||
  "https://localhost:8080";

export const useAxiosPrivate = () => {
  useEffect(() => {
    const reqIntercept = axios.interceptors.request.use(
      (config) => {
        config.withCredentials = true;
        config.baseURL = apiURL;
        return config;
      },
      (error) => Promise.reject(error),
    );

    return () => axios.interceptors.request.eject(reqIntercept);
  }, []);

  return axios;
};

export default useAxiosPrivate;
