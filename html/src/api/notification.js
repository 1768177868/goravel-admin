import request from '../utils/request'

export function fetchNotifications(params = {}) {
  return request({
    url: '/notifications',
    method: 'get',
    params
  })
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

export function createNotification(data) {
  return request({
    url: '/notifications',
    method: 'post',
    data
  })
}


