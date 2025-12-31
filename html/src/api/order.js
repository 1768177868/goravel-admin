import request from '../utils/request'

// 获取订单列表
export function getOrderList(params) {
  return request({
    url: '/orders',
    method: 'get',
    params
  })
}

// 获取订单详情
export function getOrderDetail(id, params = {}) {
  return request({
    url: `/orders/${id}`,
    method: 'get',
    params
  })
}

// 创建订单
export function createOrder(data) {
  return request({
    url: '/orders',
    method: 'post',
    data
  })
}

// 更新订单
export function updateOrder(id, data) {
  return request({
    url: `/orders/${id}`,
    method: 'put',
    data
  })
}

// 删除订单
export function deleteOrder(id) {
  return request({
    url: `/orders/${id}`,
    method: 'delete'
  })
}

// 导出订单（异步）
export function exportOrder(params) {
  return request({
    url: '/orders/export',
    method: 'post',
    data: params
  })
}

// 查询导出状态
export function getExportStatus(exportId) {
  return request({
    url: `/orders/export/status/${exportId}`,
    method: 'get'
  })
}

