import { defineStore } from 'pinia'

const detectBrowserTimezone = () => {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
  } catch {
    return 'UTC'
  }
}

export const useAppStore = defineStore('app', {
  state: () => ({
    sidebarCollapsed: localStorage.getItem('sidebarCollapsed') === 'true',
    layoutSize: localStorage.getItem('layoutSize') || 'default', // default, large, small
    isFullscreen: false,
    timezone: localStorage.getItem('timezone') || detectBrowserTimezone()
  }),

  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      localStorage.setItem('sidebarCollapsed', this.sidebarCollapsed.toString())
    },

    setSidebarCollapsed(collapsed) {
      this.sidebarCollapsed = collapsed
      localStorage.setItem('sidebarCollapsed', collapsed.toString())
    },

    setLayoutSize(size) {
      this.layoutSize = size
      localStorage.setItem('layoutSize', size)
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
      localStorage.setItem('timezone', this.timezone)
    }
  }
})

