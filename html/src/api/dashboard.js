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

// 获取每月销售数据
export function getMonthlySales() {
  return request({
    url: '/dashboard/monthly-sales',
    method: 'get'
  })
}

