import request from '../utils/request'

// 获取管理员列表
export function getAdminList(params) {
  return request({
    url: '/admins',
    method: 'get',
    params
  })
}

// 获取管理员详情
export function getAdminDetail(id) {
  return request({
    url: `/admins/${id}`,
    method: 'get'
  })
}

// 创建管理员
export function createAdmin(data) {
  return request({
    url: '/admins',
    method: 'post',
    data
  })
}

// 更新管理员
export function updateAdmin(id, data) {
  return request({
    url: `/admins/${id}`,
    method: 'put',
    data
  })
}

// 删除管理员
export function deleteAdmin(id) {
  return request({
    url: `/admins/${id}`,
    method: 'delete'
  })
}

// 导出管理员
export function exportAdmin(params) {
  return request({
    url: '/admins/export',
    method: 'get',
    params
  })
}

// 重置密码
export function resetPassword(id, data) {
  return request({
    url: `/admins/${id}/password`,
    method: 'put',
    data
  })
}

// 踢出用户（删除该用户的所有token）
export function kickOutUser(id) {
  return request({
    url: `/admins/${id}/tokens`,
    method: 'delete'
  })
}

