import request from '../utils/request'
import { normalizeListResponse } from '../utils/normalize'

export async function getNotificationList(params = {}) {
  const res = await request({
    url: '/notifications',
    method: 'get',
    params
  })

  if (!res?.data) {
    return res
  }

  return normalizeListResponse({
    ...res,
    data: {
      list: res.data.notifications || [],
      total: res.data.pagination?.total || 0,
      unread_count: res.data.unread_count || 0,
      pagination: res.data.pagination
    }
  })
}

export function fetchNotifications(params = {}) {
  return getNotificationList(params)
}

export function fetchUnreadCount() {
  return request({
    url: '/notifications/unread-count',
    method: 'get'
  })
}

export function fetchRecentNotifications(params = {}) {
  return request({
    url: '/notifications/recent',
    method: 'get',
    params
  })
}

export function markNotificationRead(id) {
  return request({
    url: `/notifications/${id}/read`,
    method: 'post'
  })
}

export function markAllNotificationsRead() {
  return request({
    url: '/notifications/read-all',
    method: 'post'
  })
}

export function createNotificationWsTicket() {
  return request({
    url: '/notifications/ws-ticket',
    method: 'post'
  })
}

export function createNotification(data) {
  return request({
    url: '/notifications',
    method: 'post',
    data
  })
}


