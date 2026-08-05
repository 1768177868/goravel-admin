import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export function getSystemInfo() {
  return request({
    url: '/monitor/system-info',
    method: 'get',
  }) as Promise<ApiResponse<Record<string, unknown>>>
}

export function createSystemInfoSSE(options: { interval?: number } = {}) {
  const { interval = 2 } = options
  return `/monitor/system-info/stream?interval=${interval}`
}
