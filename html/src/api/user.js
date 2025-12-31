import request from '../utils/request'
import { createCRUDApi, extendApi } from '../utils/apiFactory'

// 创建基础 CRUD API
const baseUserApi = createCRUDApi('users')

// 扩展 API，添加自定义方法
const userApi = extendApi(baseUserApi, {
  // 重置密码
  resetPassword: (id, data) => {
    return request({
      url: `/users/${id}/password`,
      method: 'put',
      data
    })
  },
  // 更新用户余额
  updateBalance: (id, data) => {
    return request({
      url: `/users/${id}/update-balance`,
      method: 'post',
      data
    })
  }
})

// 导出所有方法
export const {
  list: getUserList,
  detail: getUserDetail,
  create: createUser,
  update: updateUser,
  delete: deleteUser,
  resetPassword,
  updateBalance
} = userApi

