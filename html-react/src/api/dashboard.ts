import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export function getCount() {
  return request({
    url: '/dashboard/count',
    method: 'get',
  }) as Promise<ApiResponse<Record<string, unknown>>>
}

export function getWeeklyUserActivity() {
  return request({
    url: '/dashboard/weekly-user-activity',
    method: 'get',
  })
}

export function getMonthlySales() {
  return request({
    url: '/dashboard/monthly-sales',
    method: 'get',
  })
}

export function getRecentActivities() {
  return request({
    url: '/dashboard/recent-activities',
    method: 'get',
  })
}

export function getUserAccessSource() {
  return request({
    url: '/dashboard/user-access-source',
    method: 'get',
  })
}
