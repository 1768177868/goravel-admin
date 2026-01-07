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

