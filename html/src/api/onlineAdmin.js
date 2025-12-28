import request from '../utils/request'

// 获取在线管理员列表
export function getOnlineAdminList(params) {
  return request({
    url: '/online-admins',
    method: 'get',
    params
  })
}

// 踢下线（删除token）
export function kickOutOnlineAdmin(id) {
  return request({
    url: `/online-admins/${id}`,
    method: 'delete'
  })
}

// 批量踢下线
export function batchKickOutOnlineAdmins(tokenIds) {
  return request({
    url: '/online-admins/batch-kick-out',
    method: 'post',
    data: {
      token_ids: tokenIds.join(',')
    }
  })
}

