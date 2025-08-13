import axios from 'axios'
import { toast } from 'sonner'

// Create axios instance
export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8088',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add timestamp to prevent caching
    if (config.method === 'get') {
      config.params = {
        ...config.params,
        _t: Date.now(),
      }
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    const { response } = error

    if (!response) {
      toast.error('Network error. Please check your connection.')
      return Promise.reject(error)
    }

    const { status, data } = response

    switch (status) {
      case 400:
        toast.error(data?.error || 'Bad request')
        break
      case 401:
        toast.error('Authentication required')
        // Clear auth state if token is invalid
        localStorage.removeItem('evalforge-auth')
        window.location.href = '/login'
        break
      case 403:
        toast.error('Access denied')
        break
      case 404:
        toast.error('Resource not found')
        break
      case 429:
        toast.error('Too many requests. Please try again later.')
        break
      case 500:
        toast.error('Server error. Please try again later.')
        break
      default:
        toast.error(data?.error || 'An unexpected error occurred')
    }

    return Promise.reject(error)
  }
)

export default api