import axios from 'axios';

const apiURL = import.meta.env.VITE_API_URL || 'https://localhost:8080';

export default axios.create({
    baseURL: apiURL,
    headers: { 'Content-Type': 'application/json' },
})
