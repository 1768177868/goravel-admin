import { createCRUDApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'

const paymentMethodApi = createCRUDApi('payment-methods')

export async function getPaymentMethodList(params?: Record<string, unknown>) {
  const res = await paymentMethodApi.list(params)
  return normalizeListResponse(res)
}

export const {
  detail: getPaymentMethodDetail,
  create: createPaymentMethod,
  update: updatePaymentMethod,
  delete: deletePaymentMethod,
} = paymentMethodApi
