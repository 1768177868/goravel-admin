import request from '../utils/request'
import { normalizeListResponse } from '../utils/normalize'

export async function getUserBalanceLogList(params) {
  const res = await request({
    url: '/user-balance-logs',
    method: 'get',
    params
  })
  return normalizeListResponse(res)
}

export function getUserBalanceStatistics(params) {
  return request({
    url: '/user-balance-logs/statistics',
    method: 'get',
    params
  })
}
