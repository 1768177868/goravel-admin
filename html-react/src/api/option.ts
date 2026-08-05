import request from '@/utils/request'
import type { ApiResponse } from '@/types'

export interface OptionItem {
  label?: string
  value?: string | number
  id?: string | number
  name?: string
  children?: OptionItem[]
  [key: string]: unknown
}

export function getOptions(type: string, params: Record<string, unknown> = {}) {
  return request({
    url: '/options',
    method: 'get',
    params: { type, ...params },
  }) as Promise<ApiResponse<OptionItem[] | { list?: OptionItem[]; options?: OptionItem[] }>>
}
