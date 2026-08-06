import request from '@/utils/request'
import { createCRUDApi, extendApi } from '@/utils/apiFactory'
import { normalizeListResponse } from '@/utils/normalize'

const baseUserApi = createCRUDApi('users')

const userApi = extendApi(baseUserApi, {
  resetPassword: (id: string | number, data: { password: string }) =>
    request({
      url: `/users/${id}/password`,
      method: 'put',
      data,
    }),
  updateBalance: (id: string | number, data: Record<string, unknown>) =>
    request({
      url: `/users/${id}/update-balance`,
      method: 'post',
      data,
    }),
  export: (params?: Record<string, unknown>) =>
    request({
      url: '/users/export',
      method: 'post',
      data: params,
    }),
})

export async function getUserList(params?: Record<string, unknown>) {
  const res = await userApi.list(params)
  return normalizeListResponse(res)
}

export const {
  detail: getUserDetail,
  create: createUser,
  update: updateUser,
  delete: deleteUser,
  resetPassword: resetUserPassword,
  updateBalance,
  export: exportUsers,
} = userApi
