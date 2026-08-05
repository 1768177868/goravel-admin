import { normalizeListResponse } from '@/utils/normalize'
import request from '@/utils/request'
import type { ApiResponse, PaginatedData } from '@/types'

export async function getOnlineAdminList(params?: Record<string, unknown>) {
  const res = await request({
    url: '/online-admins',
    method: 'get',
    params,
  })
  return normalizeListResponse(res) as ApiResponse<PaginatedData>
}

export function kickOutOnlineAdmin(id: string | number) {
  return request({
    url: `/online-admins/${id}`,
    method: 'delete',
  })
}

export function batchKickOutOnlineAdmins(tokenIds: Array<string | number>) {
  return request({
    url: '/online-admins/batch-kick-out',
    method: 'post',
    data: { token_ids: tokenIds.join(',') },
  })
}
