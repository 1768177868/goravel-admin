import request from '@/utils/request'
import type { ApiResponse, CaptchaInfo, LoginPayload, UserInfoPayload } from '@/types'

export function login(data: LoginPayload) {
  return request({
    url: '/login',
    method: 'post',
    data,
  }) as Promise<ApiResponse<{ token?: string; admin?: unknown }>>
}

export function getInfo() {
  return request({
    url: '/info',
    method: 'get',
  }) as Promise<ApiResponse<UserInfoPayload>>
}

export function logout() {
  return request({
    url: '/logout',
    method: 'post',
  }) as Promise<ApiResponse<unknown>>
}

export function getLoginCaptcha() {
  return request({
    url: '/login/captcha',
    method: 'get',
  }) as Promise<ApiResponse<CaptchaInfo>>
}

export function getGoogleAuthenticatorStatus() {
  return request({
    url: '/google-authenticator/status',
    method: 'get',
  })
}

export function getGoogleAuthenticatorQRCode() {
  return request({
    url: '/google-authenticator/qrcode',
    method: 'get',
  })
}

export function bindGoogleAuthenticator(data: { code: string }) {
  return request({
    url: '/google-authenticator/bind',
    method: 'post',
    data,
  })
}

export function unbindGoogleAuthenticator(data: { code: string }) {
  return request({
    url: '/google-authenticator/unbind',
    method: 'post',
    data,
  })
}
