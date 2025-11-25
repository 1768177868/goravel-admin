import request from '../utils/request'

export function getSystemInfo() {
  return request({
    url: '/monitor/system-info',
    method: 'get'
  })
}

