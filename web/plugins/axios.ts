import axios from 'axios'
import type { AxiosInstance } from 'axios'

let client: AxiosInstance | null = null

function createClient(baseURL: string): AxiosInstance {
  client = axios.create({
    baseURL,
    withCredentials: true,
    headers: {
      'X-Client-Version': 'dev',
    },
  })
  return client
}

export function getAxios(): AxiosInstance {
  if (!client) {
    client = createClient('http://localhost:8000')
  }
  return client
}

export function setAuthHeader(token: string | null) {
  if (token) {
    getAxios().defaults.headers.common.Authorization = `Bearer ${token}`
  } else {
    delete getAxios().defaults.headers.common.Authorization
  }
}

export function setApiKey(apiKey: string | null) {
  if (apiKey) {
    getAxios().defaults.headers.common['x-api-key'] = apiKey
  } else {
    delete getAxios().defaults.headers.common['x-api-key']
  }
}

export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig()
  createClient(config.public.apiBaseUrl as string)

  getAxios().interceptors.response.use(
    (response) => response,
    (error) => {
      if (error?.response?.status === 401) {
        const store = useAppStore()
        store.clearAuthUser()
        const route = useRoute()
        if (route.path !== '/login') {
          navigateTo(`/login?to=${route.path}`)
        }
      }
      return Promise.reject(error)
    },
  )
})
