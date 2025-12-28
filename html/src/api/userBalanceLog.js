import request from '../utils/request'

// 获取用户余额变动记录列表
export const getUserBalanceLogList = (params) => {
  return request({
    url: '/user-balance-logs',
    method: 'get',
    params
  })
}

// 获取用户余额统计
export const getUserBalanceStatistics = (params) => {
  return request({
    url: '/user-balance-logs/statistics',
    method: 'get',
    params
  })
}

