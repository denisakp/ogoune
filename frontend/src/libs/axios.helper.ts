import axios, { AxiosError } from 'axios'
import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import { message } from 'ant-design-vue'

const SUCCESS_TOAST_METHODS = ['post', 'put', 'patch', 'delete']

interface CustomRequestConfig {
  skipSuccessToast?: boolean
  skipErrorToast?: boolean
  successMessage?: string
}

interface RequestConfig extends InternalAxiosRequestConfig, CustomRequestConfig {}

// Create axios instance with default config
const axiosClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor - allows configuring toast options per request
axiosClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig & CustomRequestConfig) => {
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  },
)

// Response interceptor for success handling (2xx and 3xx)
axiosClient.interceptors.response.use(
  (response: AxiosResponse) => {
    const config = response.config as InternalAxiosRequestConfig & CustomRequestConfig
    const method = config.method?.toLowerCase()

    // Show success toast for modification operations (POST, PUT, PATCH, DELETE)
    // unless explicitly disabled
    if (method && SUCCESS_TOAST_METHODS.includes(method) && !config.skipSuccessToast) {
      const successMessage = config.successMessage || getDefaultSuccessMessage(method)
      message.success(successMessage)
    }

    return response
  },
  (error: AxiosError) => {
    const config = error.config as (InternalAxiosRequestConfig & CustomRequestConfig) | undefined

    // Don't show error toast if explicitly disabled
    if (config?.skipErrorToast) {
      return Promise.reject(error)
    }

    // Handle HTTP errors (4xx and 5xx)
    if (error.response) {
      const status = error.response.status
      const errorMessage = extractErrorMessage(error)

      // Client errors (4xx)
      if (status >= 400 && status < 500) {
        switch (status) {
          case 400:
            message.error(`Bad request: ${errorMessage}`)
            break
          case 401:
            message.error('Unauthorized. Please log in again.')
            break
          case 403:
            message.error('Access forbidden.')
            break
          case 404:
            message.error('Resource not found.')
            break
          case 409:
            message.error(`Conflict: ${errorMessage}`)
            break
          case 422:
            message.error(`Validation failed: ${errorMessage}`)
            break
          default:
            message.error(`Client error: ${errorMessage}`)
        }
      }
      // Server errors (5xx)
      else if (status >= 500) {
        switch (status) {
          case 500:
            message.error(`Server error: ${errorMessage}`)
            break
          case 502:
            message.error('Service temporarily unavailable.')
            break
          case 503:
            message.error('Service under maintenance.')
            break
          case 504:
            message.error('Request timeout.')
            break
          default:
            message.error(`Server error: ${errorMessage}`)
        }
      }
    }
    // Network errors (no response from server)
    else if (error.request) {
      message.error('Unable to reach the server. Please check your connection.')
    }
    // Configuration errors
    else {
      message.error(`Error: ${error.message}`)
    }

    console.error('API Error:', error.message, error.response?.data)
    return Promise.reject(error)
  },
)

/**
 * Extract error message from API response
 */
function extractErrorMessage(error: AxiosError): string {
  if (error.response?.data) {
    const data = error.response.data as Record<string, unknown>
    // Try different common error response structures
    return (
      (data.message as string) ||
      (data.error as string) ||
      (data.detail as string) ||
      (data.msg as string) ||
      error.message ||
      'An error occurred'
    )
  }
  return error.message || 'An error occurred'
}

/**
 * Generate default success message based on HTTP method
 */
function getDefaultSuccessMessage(method: string): string {
  switch (method) {
    case 'post':
      return 'Created successfully'
    case 'put':
    case 'patch':
      return 'Updated successfully'
    case 'delete':
      return 'Deleted successfully'
    default:
      return 'Operation successful'
  }
}

export default axiosClient

// Export types to allow custom configuration
export type { CustomRequestConfig }
