import request from '../utils/request'

// 获取黑名单列表
export function getBlacklistList(params) {
  return request({
    url: '/blacklists',
    method: 'get',
    params
  })
}

// 获取黑名单详情
export function getBlacklistDetail(id) {
  return request({
    url: `/blacklists/${id}`,
    method: 'get'
  })
}

// 创建黑名单
export function createBlacklist(data) {
  return request({
    url: '/blacklists',
    method: 'post',
    data
  })
}

// 更新黑名单
export function updateBlacklist(id, data) {
  return request({
    url: `/blacklists/${id}`,
    method: 'put',
    data
  })
}

// 删除黑名单
export function deleteBlacklist(id) {
  return request({
    url: `/blacklists/${id}`,
    method: 'delete'
  })
}

