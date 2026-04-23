import axios, { AxiosError, type AxiosRequestConfig } from 'axios'
import type { ApiResponse, ApiError, Paginated } from '@/types/common'

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
  baseURL: '/api',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

instance.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiResponse<unknown>>) => {
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
  const res = await instance.get<ApiResponse<T[]>>(url, config)
  const data = unwrap(res.data, res.status)
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

export { instance as httpClient }
