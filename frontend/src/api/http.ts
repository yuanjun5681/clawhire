import axios, { AxiosError, type AxiosRequestConfig } from 'axios'
import type { ApiResponse, ApiError, Paginated } from '@/types/common'
import { API_BASE_URL } from './config'
import { loadSessionSnapshot } from '@/stores/identity'

export class ApiRequestError extends Error {
  code: string
  status?: number

  constructor(error: ApiError, status?: number) {
    super(error.message)
    this.name = 'ApiRequestError'
    this.code = error.code
    this.status = status
  }
}

const instance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

instance.interceptors.request.use((config) => {
  const session = loadSessionSnapshot()
  if (session?.token) {
    config.headers = config.headers ?? {}
    config.headers['Authorization'] = `Bearer ${session.token}`
  }
  return config
})

// 401 处理器：避免 store 循环依赖，这里通过事件解耦。
export const UNAUTHORIZED_EVENT = 'clawhire:unauthorized'

instance.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiResponse<unknown>>) => {
    if (error.response?.status === 401) {
      window.dispatchEvent(new CustomEvent(UNAUTHORIZED_EVENT))
    }
    const payload = error.response?.data
    if (payload && payload.success === false) {
      return Promise.reject(
        new ApiRequestError(payload.error, error.response?.status),
      )
    }
    return Promise.reject(
      new ApiRequestError(
        {
          code: 'NETWORK_ERROR',
          message: error.message || '请求失败',
        },
        error.response?.status,
      ),
    )
  },
)

function unwrap<T>(payload: ApiResponse<T>, status?: number): T {
  if (payload.success) return payload.data
  throw new ApiRequestError(payload.error, status)
}

export async function httpGet<T>(
  url: string,
  config?: AxiosRequestConfig,
): Promise<T> {
  const res = await instance.get<ApiResponse<T>>(url, config)
  return unwrap(res.data, res.status)
}

export async function httpGetPaginated<T>(
  url: string,
  config?: AxiosRequestConfig,
): Promise<Paginated<T>> {
  const res = await instance.get<ApiResponse<T[] | null>>(url, config)
  const data = unwrap(res.data, res.status) ?? []
  const meta = res.data.success ? res.data.meta ?? {} : {}
  return {
    items: data,
    page: meta.page ?? 1,
    pageSize: meta.pageSize ?? data.length,
    total: meta.total ?? data.length,
  }
}

export async function httpPost<TReq, TRes>(
  url: string,
  body?: TReq,
  config?: AxiosRequestConfig,
): Promise<TRes> {
  const res = await instance.post<ApiResponse<TRes>>(url, body, config)
  return unwrap(res.data, res.status)
}

export async function httpDelete<TRes = void>(
  url: string,
  config?: AxiosRequestConfig,
): Promise<TRes> {
  const res = await instance.delete<ApiResponse<TRes>>(url, config)
  return unwrap(res.data, res.status)
}

export { instance as httpClient }
