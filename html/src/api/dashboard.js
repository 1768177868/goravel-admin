import request from '../utils/request'

// 获取统计数据
export function getCount() {
  return request({
    url: '/dashboard/count',
    method: 'get'
  })
}

// 获取用户访问来源
export function getUserAccessSource() {
  return request({
    url: '/dashboard/user-access-source',
    method: 'get'
  })
}

// 获取每周用户活动
export function getWeeklyUserActivity() {
  return request({
    url: '/dashboard/weekly-user-activity',
    method: 'get'
  })
}

// 获取每月销售数据（实际是操作统计）
export function getMonthlySales() {
  return request({
    url: '/dashboard/monthly-sales',
    method: 'get'
  })
}

// 获取最近活动
export function getRecentActivities() {
  return request({
    url: '/dashboard/recent-activities',
    method: 'get'
  })
}

// 创建 Dashboard 数据实时更新 SSE URL
export function createDashboardSSE(options = {}) {
  const { interval = 5 } = options
  const url = `/dashboard/stream?interval=${interval}`
  return url
}

