import request from '../utils/request'

// 获取支付记录列表
export function getPaymentList(params) {
  return request({
    url: '/payments',
    method: 'get',
    params
  })
}

// 获取支付记录详情
export function getPaymentDetail(id) {
  return request({
    url: `/payments/${id}`,
    method: 'get'
  })
}

// 导出支付记录
export function exportPayments(params) {
  return request({
    url: '/payments/export',
    method: 'post',
    params
  })
}

// 获取导出状态
export function getExportStatus(id) {
  return request({
    url: `/payments/export/status/${id}`,
    method: 'get'
  })
}
