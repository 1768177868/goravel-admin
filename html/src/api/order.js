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
// 支持通过订单ID或订单号查询
// 如果提供了order_no参数，优先使用订单号查询（更高效，可直接定位分表）
// 如果没有order_no，则使用id参数查询
export function getOrderDetail(id, params = {}) {
  // 如果params中有order_no，优先使用订单号查询
  if (params.order_no) {
    return request({
      url: `/orders/${id || 0}`, // id可以为0，因为会使用order_no查询
      method: 'get',
      params: {
        order_no: params.order_no
      }
    })
  }
  // 否则使用订单ID查询
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
// 如果data中有order_no，优先使用订单号查询（更高效，可直接定位分表）
export function updateOrder(id, data) {
  const params = {}
  if (data.order_no) {
    params.order_no = data.order_no
  }
  return request({
    url: `/orders/${id || 0}`, // id可以为0，因为会使用order_no查询
    method: 'put',
    data,
    params
  })
}

// 删除订单
// 如果提供了order_no参数，优先使用订单号查询（更高效，可直接定位分表）
export function deleteOrder(id, params = {}) {
  return request({
    url: `/orders/${id || 0}`, // id可以为0，因为会使用order_no查询
    method: 'delete',
    params
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

// 导入订单
export function importOrder(file) {
  const formData = new FormData()
  formData.append('file', file)
  return request({
    url: '/orders/import',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

