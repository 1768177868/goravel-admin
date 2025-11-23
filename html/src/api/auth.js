import request from '../utils/request'

// 登录
export function login(data) {
  return request({
    url: '/login',
    method: 'post',
    data
  })
}

// 获取当前用户信息
export function getInfo() {
  return request({
    url: '/info',
    method: 'get'
  })
}

// 退出登录
export function logout() {
  return request({
    url: '/logout',
    method: 'post'
  })
}

// 获取 token 列表
export function getTokens() {
  return request({
    url: '/tokens',
    method: 'get'
  })
}

// 删除指定 token
export function revokeToken(id) {
  return request({
    url: `/tokens/${id}`,
    method: 'delete'
  })
}

// 删除所有 token
export function revokeAllTokens() {
  return request({
    url: '/tokens',
    method: 'delete'
  })
}

