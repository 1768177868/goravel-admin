import { defineStore } from 'pinia'
import Storage from '../utils/storage'

const detectBrowserTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

export const useAppStore = defineStore('app', {
  state: () => ({
    sidebarCollapsed: Storage.getItem('sidebarCollapsed', 'false') === 'true',
    layoutSize: Storage.getItem('layoutSize', 'default') || 'default', // default, large, small
    isFullscreen: false,
    timezone: Storage.getItem('timezone', detectBrowserTimezone()) || detectBrowserTimezone()
  }),

  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      Storage.setItem('sidebarCollapsed', this.sidebarCollapsed.toString())
    },

    setSidebarCollapsed(collapsed) {
      this.sidebarCollapsed = collapsed
      Storage.setItem('sidebarCollapsed', collapsed.toString())
    },

    setLayoutSize(size) {
      this.layoutSize = size
      Storage.setItem('layoutSize', size)
      // 应用布局大小到 body
      document.body.className = document.body.className.replace(/layout-\w+/g, '')
      document.body.classList.add(`layout-${size}`)
    },

    toggleFullscreen() {
      if (!document.fullscreenElement) {
        document.documentElement.requestFullscreen().then(() => {
          this.isFullscreen = true
        }).catch(() => {
          console.error('无法进入全屏模式')
        })
      } else {
        document.exitFullscreen().then(() => {
          this.isFullscreen = false
        }).catch(() => {
          console.error('无法退出全屏模式')
        })
      }
    },

    setTimezone(timezone) {
      this.timezone = timezone || 'UTC'
      Storage.setItem('timezone', this.timezone)
    }
  }
})

