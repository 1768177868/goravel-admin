import request from '../utils/request'

export function getSystemInfo() {
  return request({
    url: '/monitor/system-info',
    method: 'get'
  })
}

// 创建系统监控实时数据流 SSE URL
export function createSystemInfoSSE(options = {}) {
  const { interval = 2 } = options
  const url = `/monitor/system-info/stream?interval=${interval}`
  return url
}

