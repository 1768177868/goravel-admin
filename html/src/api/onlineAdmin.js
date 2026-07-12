import request from '../utils/request'
import { normalizeListResponse } from '../utils/normalize'

export async function getOnlineAdminList(params) {
  const res = await request({
    url: '/online-admins',
    method: 'get',
    params
  })
  return normalizeListResponse(res)
}

export function kickOutOnlineAdmin(id) {
  return request({
    url: `/online-admins/${id}`,
    method: 'delete'
  })
}

export function batchKickOutOnlineAdmins(tokenIds) {
  return request({
    url: '/online-admins/batch-kick-out',
    method: 'post',
    data: {
      token_ids: tokenIds.join(',')
    }
  })
}
