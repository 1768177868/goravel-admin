import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export function getConfigByGroup(group: string) {
  return request({
    url: `/configs/group/${group}`,
    method: 'get',
  }) as Promise<ApiResponse<{ configs?: Array<Record<string, unknown>> }>>
}

export function saveConfig(group: string, configs: Record<string, unknown>) {
  return request({
    url: '/configs/save',
    method: 'post',
    data: { group, configs },
  }) as Promise<ApiResponse<unknown>>
}

export function testEmail(emailConfig: Record<string, unknown>) {
  return request({
    url: '/configs/test-email',
    method: 'post',
    data: emailConfig,
  }) as Promise<ApiResponse<unknown>>
}
