import axios from 'axios';

const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export default axios.create({
    baseURL: apiURL,
    headers: {'Content-Type': 'application/json'},
})