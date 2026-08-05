import request from '@/utils/request'
import { normalizeListResponse } from '@/utils/normalize'
import type { ApiResponse, PaginatedData } from '@/types'

export async function getNotificationList(params: Record<string, unknown> = {}) {
  const res = await request({
    url: '/notifications',
    method: 'get',
    params,
  })

  if (!res?.data) return res as ApiResponse<PaginatedData>

  return normalizeListResponse({
    ...res,
    data: {
      list: (res.data as { notifications?: unknown[] }).notifications || [],
      total: (res.data as { pagination?: { total?: number } }).pagination?.total || 0,
      unread_count: (res.data as { unread_count?: number }).unread_count || 0,
      pagination: (res.data as { pagination?: unknown }).pagination,
    },
  }) as ApiResponse<PaginatedData & { unread_count?: number }>
}

export function fetchUnreadCount() {
  return request({ url: '/notifications/unread-count', method: 'get' })
}

export function fetchRecentNotifications(params: Record<string, unknown> = {}) {
  return request({ url: '/notifications/recent', method: 'get', params })
}

export function markNotificationRead(id: string | number) {
  return request({ url: `/notifications/${id}/read`, method: 'post' })
}

export function markAllNotificationsRead() {
  return request({ url: '/notifications/read-all', method: 'post' })
}

export function createNotificationWsTicket() {
  return request({ url: '/notifications/ws-ticket', method: 'post' })
}

export function createNotification(data: Record<string, unknown>) {
  return request({ url: '/notifications', method: 'post', data })
}
