import { defineStore } from 'pinia'
import { ElMessage } from 'element-plus'
import i18n from '../i18n'
import {
  fetchNotifications,
  fetchUnreadCount,
  fetchRecentNotifications,
  markNotificationRead,
  markAllNotificationsRead
} from '../api/notification'

const { t } = i18n.global

export const useNotificationStore = defineStore('notification', {
  state: () => ({
    items: [],
    unreadCount: 0,
    loading: false,
    ws: null,
    wsConnected: false,
    initializing: false,
    retryCount: 0,
    retryTimer: null
  }),
  actions: {
    async init() {
      if (this.initializing) {
        return
      }
      this.initializing = true
      await this.refresh()
      this.connect()
    },
    async refresh(params = {}) {
      this.loading = true
      try {
        const { data } = await fetchRecentNotifications({
          limit: params.limit || 7
        })
        this.items = data.notifications || []
        this.unreadCount = data.unread_count || 0
      } catch (error) {
        console.error('Load notifications error:', error)
      } finally {
        this.loading = false
      }
    },
    async fetchUnread() {
      try {
        const { data } = await fetchUnreadCount()
        this.unreadCount = data.count || 0
      } catch (error) {
        console.error('Fetch unread count error:', error)
      }
    },
    async markAsRead(id) {
      try {
        await markNotificationRead(id)
        this.items = this.items.map(item =>
          item.id === id ? { ...item, is_read: true, read_at: new Date().toISOString() } : item
        )
        if (this.unreadCount > 0) {
          this.unreadCount -= 1
        }
      } catch (error) {
        console.error('Mark notification read failed:', error)
      }
    },
    async markAllRead() {
      try {
        await markAllNotificationsRead()
        this.items = this.items.map(item => ({ ...item, is_read: true, read_at: new Date().toISOString() }))
        this.unreadCount = 0
      } catch (error) {
        console.error('Mark all notifications read failed:', error)
      }
    },
    connect() {
      if (this.ws || this.wsConnected) {
        return
      }
      const token = localStorage.getItem('token')
      if (!token) {
        return
      }
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      const wsUrl = `${protocol}//${host}/ws/admin/notifications?token=${encodeURIComponent(token.trim())}`

      this.ws = new WebSocket(wsUrl)
      this.ws.onopen = () => {
        this.wsConnected = true
        this.retryCount = 0
      }
      this.ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data)
          this.handleIncoming(payload)
        } catch (error) {
          console.error('Invalid notification payload:', error)
        }
      }
      this.ws.onclose = () => {
        this.wsConnected = false
        this.ws = null
        this.scheduleReconnect()
      }
      this.ws.onerror = () => {
        this.wsConnected = false
      }
    },
    scheduleReconnect() {
      if (!this.retryCount) {
        this.retryCount = 0
      }
      if (this.retryTimer) {
        clearTimeout(this.retryTimer)
      }
      const delay = Math.min(30000, 2000 * Math.pow(2, this.retryCount))
      this.retryTimer = setTimeout(() => {
        this.retryCount += 1
        const token = localStorage.getItem('token')
        if (!token) {
          return
        }
        this.connect()
        if (!this.wsConnected && this.retryCount === 3) {
          ElMessage.error(t('notification.ws_error'))
        }
      }, delay)
    },
    disconnect() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
      this.wsConnected = false
      this.initializing = false
      this.items = []
      this.unreadCount = 0
      if (this.retryTimer) {
        clearTimeout(this.retryTimer)
        this.retryTimer = null
      }
      this.retryCount = 0
    },
    handleIncoming(notification) {
      const exists = this.items.find(item => item.id === notification.id)
      if (!exists) {
        this.items.unshift(notification)
        if (!notification.is_read) {
          this.unreadCount += 1
        }
        if (this.items.length > 7) {
          this.items = this.items.slice(0, 7)
        }
      } else {
        this.items = this.items.map(item => item.id === notification.id ? notification : item)
      }
    }
  }
})

