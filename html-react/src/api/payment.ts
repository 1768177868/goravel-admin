import request from '@/utils/request'
import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeEntity, normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const paymentApi = extendApi(createCRUDApi('payments'), {
  export: (params?: Record<string, unknown>) =>
    request({
      url: '/payments/export',
      method: 'post',
      data: params,
    }),
  exportStatus: (id: string | number) =>
    request({
      url: `/payments/export/status/${id}`,
      method: 'get',
    }),
})

export async function getPaymentList(params?: Record<string, unknown>) {
  const res = await paymentApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export async function getPaymentDetail(id: string | number) {
  const res = await paymentApi.detail(id)
  if (res?.data && typeof res.data === 'object') {
    const data = res.data as Record<string, unknown>
    const payload = data.data || data
    if (payload && typeof payload === 'object') {
      const normalized = normalizeEntity(payload as Record<string, unknown>)
      if (data.data) {
        data.data = normalized
      } else {
        Object.assign(data, normalized)
      }
    }
  }
  return res
}

export const exportPayments = paymentApi.export
export const getExportStatus = paymentApi.exportStatus
