import request from './request'
import type { ApiResponse } from '@/types'

export interface CrudApi<T = unknown> {
  list: (params?: Record<string, unknown>) => Promise<ApiResponse<{ list?: T[]; data?: T[]; total?: number }>>
  detail: (id: string | number) => Promise<ApiResponse<T>>
  create: (data: unknown) => Promise<ApiResponse<T>>
  update: (id: string | number, data: unknown) => Promise<ApiResponse<T>>
  delete: (id: string | number) => Promise<ApiResponse<unknown>>
  batchDelete: (ids: Array<string | number>) => Promise<ApiResponse<unknown>>
}

export function createCRUDApi<T = unknown>(resource: string): CrudApi<T> {
  return {
    list: (params) =>
      request({
        url: `/${resource}`,
        method: 'get',
        params,
      }) as Promise<ApiResponse<{ list?: T[]; data?: T[]; total?: number }>>,
    detail: (id) =>
      request({
        url: `/${resource}/${id}`,
        method: 'get',
      }) as Promise<ApiResponse<T>>,
    create: (data) =>
      request({
        url: `/${resource}`,
        method: 'post',
        data,
      }) as Promise<ApiResponse<T>>,
    update: (id, data) =>
      request({
        url: `/${resource}/${id}`,
        method: 'put',
        data,
      }) as Promise<ApiResponse<T>>,
    delete: (id) =>
      request({
        url: `/${resource}/${id}`,
        method: 'delete',
      }) as Promise<ApiResponse<unknown>>,
    batchDelete: (ids) =>
      request({
        url: `/${resource}/batch`,
        method: 'delete',
        data: { ids },
      }) as Promise<ApiResponse<unknown>>,
  }
}

export function extendApi<TBase extends object, TExtra extends object>(
  baseApi: TBase,
  customMethods: TExtra,
): TBase & TExtra {
  return {
    ...baseApi,
    ...customMethods,
  }
}
