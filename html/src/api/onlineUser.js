import request from '../utils/request'

// 获取在线用户列表
export function getOnlineUserList(params) {
  return request({
    url: '/online-users',
    method: 'get',
    params
  })
}

// 踢下线（删除token）
export function kickOutOnlineUser(id) {
  return request({
    url: `/online-users/${id}`,
    method: 'delete'
  })
}

// 批量踢下线
export function batchKickOutOnlineUsers(tokenIds) {
  return request({
    url: '/online-users/batch-kick-out',
    method: 'post',
    data: {
      token_ids: tokenIds.join(',')
    }
  })
}

