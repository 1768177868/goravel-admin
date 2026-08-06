import request from '@/utils/request'
import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse } from '@/types'

const baseOrderApi = createCRUDApi('orders')

const orderApi = extendApi(baseOrderApi, {
  export: (params?: Record<string, unknown>) =>
    request({
      url: '/orders/export',
      method: 'post',
      data: params,
    }),
  exportStatus: (exportId: string | number) =>
    request({
      url: `/orders/export/status/${exportId}`,
      method: 'get',
    }),
  import: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request({
      url: '/orders/import',
      method: 'post',
      data: formData,
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },
})

export async function getOrderList(params?: Record<string, unknown>) {
  const res = await orderApi.list(params)
  return normalizeListResponse(res)
}

export function getOrderDetail(id: string | number, params: Record<string, unknown> = {}) {
  if (params.order_no) {
    return request({
      url: `/orders/${id || 0}`,
      method: 'get',
      params: { order_no: params.order_no },
    }) as Promise<ApiResponse<unknown>>
  }
  return orderApi.detail(id)
}

export function createOrder(data: unknown) {
  return orderApi.create(data)
}

export function updateOrder(id: string | number, data: unknown) {
  return orderApi.update(id, data)
}

export function deleteOrder(id: string | number, params: Record<string, unknown> = {}) {
  return request({
    url: `/orders/${id}`,
    method: 'delete',
    params: Object.keys(params).length ? params : undefined,
  }) as Promise<ApiResponse<unknown>>
}

export const exportOrder = orderApi.export
export const getExportStatus = orderApi.exportStatus
export const importOrder = orderApi.import
