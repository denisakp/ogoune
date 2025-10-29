import axios, { AxiosError } from 'axios'
import type { AxiosInstance } from 'axios'

//create axios instance with default config

const axiosClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// add response interceptor for error handling
axiosClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    console.error('API Error:', error.message)
    return Promise.reject(error)
  },
)

export default axiosClient
