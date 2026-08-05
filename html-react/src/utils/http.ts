import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios'
import type { ApiResponse } from '@/types'
import { getApiBaseURL } from './env'

export type RequestConfig = AxiosRequestConfig & {
  skipErrorMessage?: boolean
  silent?: boolean
}

/** Underlying axios instance (interceptors attach here). */
export const http: AxiosInstance = axios.create({
  baseURL: getApiBaseURL(),
  timeout: 60000,
})

export type TypedRequest = {
  <T = unknown>(config: RequestConfig): Promise<ApiResponse<T>>
  get<T = unknown>(url: string, config?: RequestConfig): Promise<ApiResponse<T>>
  post<T = unknown>(url: string, data?: unknown, config?: RequestConfig): Promise<ApiResponse<T>>
  put<T = unknown>(url: string, data?: unknown, config?: RequestConfig): Promise<ApiResponse<T>>
  delete<T = unknown>(url: string, config?: RequestConfig): Promise<ApiResponse<T>>
}

export default http as TypedRequest & AxiosInstance
