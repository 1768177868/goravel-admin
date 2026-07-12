import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'
import { normalizeListResponse } from '../utils/normalize'

const baseUserApi = createCRUDApi('users')

const userApi = extendApi(baseUserApi, {
  resetPassword: (id, data) => {
    return request({
      url: `/users/${id}/password`,
      method: 'put',
      data
    })
  },
  updateBalance: (id, data) => {
    return request({
      url: `/users/${id}/update-balance`,
      method: 'post',
      data
    })
  },
  export: (params) => {
    return request({
      url: '/users/export',
      method: 'post',
      data: params
    })
  }
})

export async function getUserList(params) {
  const res = await userApi.list(params)
  return normalizeListResponse(res)
}

export const {
  detail: getUserDetail,
  create: createUser,
  update: updateUser,
  delete: deleteUser,
  resetPassword,
  updateBalance
} = userApi

export const exportUsers = userApi.export
