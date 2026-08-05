import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

const paymentMethodApi = createCRUDApi('payment-methods')

export async function getPaymentMethodList(params?: Record<string, unknown>) {
  const res = await paymentMethodApi.list(params)
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export const {
  detail: getPaymentMethodDetail,
  create: createPaymentMethod,
  update: updatePaymentMethod,
  delete: deletePaymentMethod,
} = paymentMethodApi
