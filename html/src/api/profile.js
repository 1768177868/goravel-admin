import request from '../utils/request'

// 获取个人信息
export function getProfile() {
  return request({
    url: '/info',
    method: 'get'
  })
}

// 更新个人信息
export function updateProfile(data) {
  return request({
    url: '/profile',
    method: 'put',
    data
  })
}

// 修改密码
export function updatePassword(data) {
  return request({
    url: '/password',
    method: 'put',
    data
  })
}

