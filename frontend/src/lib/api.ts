import axios from 'axios'

export const api = axios.create({ baseURL: '/api', timeout: 30000 })

api.interceptors.request.use((config) => {
  const apiKey = localStorage.getItem('iv_api_key')
  const sessionToken = localStorage.getItem('iv_session_token')
  const developerToken = localStorage.getItem('iv_developer_token')
  const adminToken = localStorage.getItem('iv_admin_token')
  if (apiKey) config.headers['X-API-Key'] = apiKey
  if (sessionToken) config.headers['X-Session-Token'] = sessionToken
  if (adminToken) config.headers.Authorization = `Bearer ${adminToken}`
  else if (developerToken) config.headers.Authorization = `Bearer ${developerToken}`
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('iv_admin_token')
    }
    return Promise.reject(error)
  },
)

export function apiMessage(error: unknown): string {
  const value = error as { response?: { data?: { error?: string; message?: string } }; message?: string }
  return value.response?.data?.error || value.response?.data?.message || value.message || 'Request failed'
}
