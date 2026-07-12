import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'
import { normalizeListResponse, normalizeEntity } from '../utils/normalize'

const basePaymentApi = createCRUDApi('payments')

const paymentApi = extendApi(basePaymentApi, {
  export: (params) => {
    return request({
      url: '/payments/export',
      method: 'post',
      data: params
    })
  },
  exportStatus: (id) => {
    return request({
      url: `/payments/export/status/${id}`,
      method: 'get'
    })
  }
})

export async function getPaymentList(params) {
  const res = await paymentApi.list(params)
  return normalizeListResponse(res)
}

export async function getPaymentDetail(id) {
  const res = await paymentApi.detail(id)
  if (res?.data) {
    const payload = res.data.data || res.data
    if (payload && typeof payload === 'object') {
      if (res.data.data) {
        res.data.data = normalizeEntity(payload)
      } else {
        Object.assign(res.data, normalizeEntity(payload))
      }
    }
  }
  return res
}

export const exportPayments = paymentApi.export
export const getExportStatus = paymentApi.exportStatus
