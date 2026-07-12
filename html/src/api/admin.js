import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

// 创建基础 CRUD API
const baseAdminApi = createCRUDApi('admins')

// 扩展 API，添加自定义方法
const adminApi = extendApi(baseAdminApi, {
  // 导出管理员
  export: (params) => {
    return request({
      url: '/admins/export',
      method: 'post',
      data: params
    })
  },

  // 重置密码
  resetPassword: (id, data) => {
    return request({
      url: `/admins/${id}/password`,
      method: 'put',
      data
    })
  },

  // 踢出用户（删除该用户的所有token）
  kickOutUser: (id) => {
    return request({
      url: `/admins/${id}/tokens`,
      method: 'delete'
    })
  },

  // 解绑管理员的谷歌验证码
  unbindGoogleAuth: (id, data) => {
    return request({
      url: `/admins/${id}/unbind-google-auth`,
      method: 'post',
      data
    })
  },

  // 重置管理员的谷歌验证码（无需验证码，用于丢失手机等场景）
  resetGoogleAuth: (id) => {
    return request({
      url: `/admins/${id}/reset-google-auth`,
      method: 'post'
    })
  }
})

export async function getAdminList(params) {
  const res = await adminApi.list(params)
  return normalizeListResponse(res)
}

// 导出所有方法（保持向后兼容）
export const {
  detail: getAdminDetail,
  create: createAdmin,
  update: updateAdmin,
  delete: deleteAdmin,
  export: exportAdmin,
  resetPassword,
  kickOutUser,
  unbindGoogleAuth: unbindAdminGoogleAuth,
  resetGoogleAuth: resetAdminGoogleAuth
} = adminApi

