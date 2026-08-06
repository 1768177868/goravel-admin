import request from '@/utils/request'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse } from '@/types'

export async function getUserBalanceLogList(params?: Record<string, unknown>) {
  const res = await request({
    url: '/user-balance-logs',
    method: 'get',
    params,
  })
  return normalizeListResponse(res)
}

export function getUserBalanceStatistics(params?: Record<string, unknown>) {
  return request({
    url: '/user-balance-logs/statistics',
    method: 'get',
    params,
  }) as Promise<
    ApiResponse<{
      total_income?: number | string
      total_expense?: number | string
      total_refund?: number | string
      current_balance?: number | string
    }>
  >
}
