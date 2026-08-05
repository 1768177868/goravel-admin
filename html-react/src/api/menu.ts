import request from '@/utils/request'
import { normalizeListResponse, normalizeTreeList } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

export async function getMenuList(params?: Record<string, unknown>) {
  const res = await request({
    url: '/menus',
    method: 'get',
    params,
  })
  return normalizeListResponse(res as unknown as ApiResponse<PaginatedData>) as ApiResponse<PaginatedData>
}

export async function getMenuTree() {
  const res = await request({
    url: '/menus/tree',
    method: 'get',
  })
  if (res?.data) {
    const list = Array.isArray(res.data) ? res.data : (res.data as { list?: unknown[] }).list
    if (Array.isArray(list)) {
      return { ...res, data: normalizeTreeList(list) }
    }
  }
  return res
}

export function getMenuDetail(id: string | number) {
  return request({ url: `/menus/${id}`, method: 'get' })
}

export function createMenu(data: unknown) {
  return request({ url: '/menus', method: 'post', data })
}

export function updateMenu(id: string | number, data: unknown) {
  return request({ url: `/menus/${id}`, method: 'put', data })
}

export function deleteMenu(id: string | number) {
  return request({ url: `/menus/${id}`, method: 'delete' })
}
