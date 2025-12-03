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

// 获取登录验证码
export function getLoginCaptcha() {
  return request({
    url: '/login/captcha',
    method: 'get'
  })
}

// 获取谷歌验证码绑定状态
export function getGoogleAuthenticatorStatus() {
  return request({
    url: '/google-authenticator/status',
    method: 'get'
  })
}

// 获取谷歌验证码二维码
export function getGoogleAuthenticatorQRCode() {
  return request({
    url: '/google-authenticator/qrcode',
    method: 'get'
  })
}

// 绑定谷歌验证码
export function bindGoogleAuthenticator(data) {
  return request({
    url: '/google-authenticator/bind',
    method: 'post',
    data
  })
}

// 解绑谷歌验证码
export function unbindGoogleAuthenticator(data) {
  return request({
    url: '/google-authenticator/unbind',
    method: 'post',
    data
  })
}

